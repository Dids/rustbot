package webrcon

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/sacOO7/gowebsocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

func (webrcon *Webrcon) handleTextMessage(message string, socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	// FIXME: Remove this
	//statusMsg := `"hostname: [FIN] Suomileijona - WIPE 7/2\nversion : 2151 secure (secure mode enabled, connected to Steam3)\nmap     : Procedural Map\nplayers : 5 (64 max) (0 queued) (0 joining)\n\nid                name                   ping connected addr                 owner violation kicks \n76561198026306491 \"NinjaMaster\"          26   58847.23s 85.76.8.230:64178          0.0       0     \n76561198162745820 \"seavaniasa\"           12   10660.97s 82.131.23.78:55801         0.0       0     \n76561198869648658 \"finjuhis90\"           19   7122.821s 80.186.198.13:53724        0.0       0     \n76561198079774759 \"Tepachu\"              5    5168.375s 84.248.190.164:55199       0.0       0     \n76561198833784648 \"John from the office\" 26   4493.813s 85.76.50.127:45012         0.0       0     \n"`
	//message = `{ "message": ` + statusMsg + `, "Identifier": 0, "Type": "Generic", "Stacktrace": "" }`

	// webrcon.logger.Trace("Received message " + message)

	// Parse the incoming message as a webrcon packet
	packet := Packet{}
	if parseErr := json.Unmarshal([]byte(message), &packet); parseErr != nil {
		webrcon.logger.Error("Failed to parse as generic message:", message, parseErr)
	}
	webrcon.logger.Trace("Parsed message as packet:", packet)

	if packet.Identifier == GenericIdentifier && packet.Type == GenericType {
		// Check if this is a valid status message
		statusRegexMatches := statusRegex.FindStringSubmatch(message)
		if len(statusRegexMatches) > 0 {
			// FIXME: It would be ideal if we could also parse the "player list" straight to the status/template
			// Template for converting status message to a JSON string
			statusTemplate := []byte(`{ "hostname": "$hostname", "version": $version, "secure": "$secure", "map": "$map", "players_current": $players_current, "players_max": $players_max, "players_queued": $players_queued, "players_joining": $players_joining }`)
			statusResult := []byte{}
			statusContent := []byte(message)

			// TODO: Isn't this sort of redundant, since we'll never have more than 1 match anyway?
			// For each match of the regex in the content
			for _, submatches := range statusRegex.FindAllSubmatchIndex(statusContent, -1) {
				// Apply the captured submatches to the template and append the output to the result
				// webrcon.logger.Trace("Result (before):\n", string(statusResult))
				statusResult = statusRegex.Expand(statusResult, statusTemplate, statusContent, submatches)
				// webrcon.logger.Trace("Result (after):\n", string(statusResult))
			}
			// webrcon.logger.Trace("End result:\n", string(statusResult))

			// Convert the resulting JSON string to a StatusPacket
			if err := json.Unmarshal(statusResult, &Status); err != nil {
				webrcon.logger.Error("Failed to parse status message:", err)
			} else {
				//webrcon.logger.Trace("Received new status:", Status)
				webrcon.logger.Trace("RAW MESSAGE:\n", message)

				// Parse the players from the status packet
				playerListRegexMatches := playerListRegex.FindStringSubmatch(message)
				webrcon.logger.Trace("playerListRegexMatches:", playerListRegexMatches)
				if len(playerListRegexMatches) > 0 {
					//if len(playerListRegexMatches)-4 > 0 {
					webrcon.logger.Trace("Matches:", len(playerListRegexMatches))
					webrcon.logger.Trace("Parsing player list..")

					// Template for converting status message to a JSON string
					playerListTemplate := []byte(`{ "steamid": "$SteamID", "username": "$Username", "ping": $Ping, "connected": "$Connected", "ip": "$IP", "port": $Port, "violations": $Violations, "kicks": $Kicks }`)
					playerListResult := []byte{}
					playerListResults := make([]*PlayerPacket, len(playerListRegexMatches))
					// playerListResults := make([]*PlayerPacket, len(playerListRegexMatches)-2 /*-4*/) // FIXME: This random "-4" here is causing issues (-2 would make more sense, right?)
					playerListContent := []byte(message)

					// For each match of the regex in the content
					for index, submatches := range playerListRegex.FindAllSubmatchIndex(playerListContent, -1) {
						// Apply the captured submatches to the template and append the output to the result
						result := playerListRegex.Expand(playerListResult, playerListTemplate, playerListContent, submatches)
						webrcon.logger.Trace("Parsing new player:\n", string(result))
						webrcon.logger.Trace(index, "/", len(playerListResults))

						if index >= len(playerListResults) {
							webrcon.logger.Error("Failed to parse player list message, index out of bounds") // FIXME: Why is this suddenly being called? What triggeres it? Our magic random numbers above?
							continue
						}

						// Convert the resulting JSON string to a list of PlayerPackets (assign to StatusPacket.Players)
						if err := json.Unmarshal(result, &playerListResults[index]); err != nil {
							webrcon.logger.Error("Failed to parse player list message:", err)
						}
					}

					// Store the new player list in Status
					Status.Players = playerListResults

					playersString, err := json.Marshal(playerListResults)
					if err != nil {
						webrcon.logger.Error("Failed to convert player list back to JSON:", err)
					} else {
						// Emit the player list change to the event handler
						webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: Status.Hostname, Message: string(playersString), Type: eventhandler.PlayersType})
					}
				} else {
					// No players online, but we still need to make sure the player list gets updated
					Status.Players = make([]*PlayerPacket, 0)
					webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: Status.Hostname, Message: "[]", Type: eventhandler.PlayersType})
				}

				// Handle message formatting depending on how many players there are
				suffix := ""

				if Status.JoiningPlayers > 0 && Status.QueuedPlayers > 0 {
					suffix = " (" + strconv.Itoa(Status.JoiningPlayers) + " joining, " + strconv.Itoa(Status.QueuedPlayers) + " queued)"
				} else if Status.JoiningPlayers > 0 {
					suffix = " (" + strconv.Itoa(Status.JoiningPlayers) + " joining)"
				} else if Status.QueuedPlayers > 0 {
					suffix = " (" + strconv.Itoa(Status.QueuedPlayers) + " queued)"
				}
				message := strconv.Itoa(Status.CurrentPlayers) + "/" + strconv.Itoa(Status.MaxPlayers) + suffix
				// webrcon.logger.Trace("Status updated, emitting status message:", message)
				webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: Status.Hostname, Message: message, Type: eventhandler.StatusType})

				return
			}
		}
	}

	//webrcon.logger.Trace("Parsed message as packet:", packet)

	// Handle different type conversions
	if packet.Identifier == ChatIdentifier && packet.Type == ChatType {
		chatPacket := ChatPacket{}
		if parseErr := json.Unmarshal([]byte(packet.Message), &chatPacket); parseErr != nil {
			webrcon.logger.Error("Failed to parse as chat message:", message, parseErr)
		}
		// webrcon.logger.Trace("Parsed message as chat packet:", chatPacket)

		// Ignore messages from "SERVER"
		if chatPacket.Username == "SERVER" {
			// webrcon.logger.Trace("NOTICE: Ignoring message from SERVER")
			return
		}

		// Send chat message to Discord
		webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: chatPacket.Username, Message: chatPacket.Message})
	} else {
		joinRegexMatches := joinRegex.FindStringSubmatch(packet.Message)
		disconnectRegexMatches := disconnectRegex.FindStringSubmatch(packet.Message)
		killRegexMatches := killRegex.FindStringSubmatch(packet.Message)
		if len(joinRegexMatches) > 1 {
			// webrcon.logger.Trace("Matched joinRegex:", joinRegexMatches)
			userID, _ := strconv.ParseUint(joinRegexMatches[3], 10, 64)
			joinPacket := JoinPacket{IP: joinRegexMatches[1], Port: joinRegexMatches[2], UserID: userID, Username: joinRegexMatches[4], OS: joinRegexMatches[5]}
			// webrcon.logger.Trace("Join packet:", joinPacket)
			webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: joinPacket.Username, Message: "joined", Type: eventhandler.JoinType})
		} else if len(disconnectRegexMatches) > 1 {
			// webrcon.logger.Trace("Matched disconnectRegex:", disconnectRegexMatches)
			userID, _ := strconv.ParseUint(disconnectRegexMatches[3], 10, 64)
			disconnectPacket := DisconnectPacket{IP: disconnectRegexMatches[1], Port: disconnectRegexMatches[2], UserID: userID, Username: disconnectRegexMatches[4]}
			// webrcon.logger.Trace("Disconnect packet:", disconnectPacket)
			webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: disconnectPacket.Username, Message: "left", Type: eventhandler.DisconnectType})
		} else if len(killRegexMatches) > 1 {
			// Construct a simple "dictionary" using the named capture groups
			result := make(map[string]string)
			for i, name := range killRegex.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = killRegexMatches[i]
				}
			}

			// Keep track of the type of kill/death
			isPvPKill := false

			// Store individual entries in local variables
			victim := result["victim"]
			victimID := result["victimid"]
			how := result["how"]
			killer := result["killer"]
			killerID := result["killerid"]
			reason := result["reason"]

			// Remove potentially leaking IDs
			victim = removeIDsRegex.ReplaceAllString(victim, "")
			killer = removeIDsRegex.ReplaceAllString(killer, "")

			// Remove words in parentheses (eg. "boar (Boar)")
			if len(killer) > 0 && len(killerID) == 0 && len(reason) > 0 && len(removeBracesRegex.FindStringSubmatch(killer)) > 1 {
				killer = strings.Replace(killer, removeBracesRegex.ReplaceAllString(killer, "$1"), "", -1)
			}
			if len(reason) > 0 && len(removeBracesRegex.FindStringSubmatch(reason)) > 1 {
				reason = strings.Replace(reason, removeBracesRegex.ReplaceAllString(reason, "$1"), "", -1)
			}

			// Reformat the death reason (eg. "X died (Bullet)")
			reason = strings.ToLower(reason)
			reason = strings.Replace(reason, "(", "", -1)
			reason = strings.Replace(reason, ")", "", -1)

			/*webrcon.logger.Trace("message:", message)
			webrcon.logger.Trace("killRegexMatches:", killRegexMatches)
			webrcon.logger.Trace("result:", result)
			webrcon.logger.Trace("victim:", victim)
			webrcon.logger.Trace("victimID:", victimID)
			webrcon.logger.Trace("how:", how)
			webrcon.logger.Trace("killer:", killer)
			webrcon.logger.Trace("killerID:", killerID)
			webrcon.logger.Trace("reason:", reason)*/

			// Skip scientist deaths
			if len(victim) > 0 && len(victimID) > 0 && victim == victimID {
				return
			}

			// Rename scientists
			if len(killer) > 0 && len(killerID) > 0 && killer == killerID {
				killer = "a scientist"
			}

			// Rename drowning
			if len(killer) > 0 && len(killerID) == 0 && len(reason) > 0 && strings.Contains(reason, "drowned") {
				// tuna was killed by drowned
				reason = "drowning"
			}

			// Construct the death message
			deathMessage := ""
			if len(victim) > 0 && len(victimID) > 0 && len(how) > 0 && len(killer) > 0 && len(killerID) > 0 && len(reason) == 0 {
				// "PlayerA was killed by PlayerB"
				deathMessage = victim + " " + how + " " + killer

				// Mark this as a PvP kill
				isPvPKill = true
			} else if len(victim) > 0 && len(victimID) > 0 && len(how) > 0 && len(reason) > 0 {
				if len(how) == 4 {
					// "PlayerA died fall"
					deathMessage = victim + " " + how + " from " + reason
				} else {
					// "PlayerA was killed by Hunger"
					deathMessage = victim + " " + how + " " + reason
				}
			} else {
				// TODO: What if our error handler DM'd us any errors? That'd be super cool and useful!
				webrcon.logger.Error("Could not parse death message:", message)
				return
			}

			// Store PvP kills/deaths in the database
			if isPvPKill {
				// FIXME: How and where do we actually update the user data, like username and such?

				// Increment the kill count for the killer
				if err := incrementKillCount(webrcon.Database, killerID); err != nil {
					webrcon.logger.Error("Failed to increment kill count: ", err)
				}

				// Increment the death count for the victim
				if err := incrementDeathCount(webrcon.Database, victimID); err != nil {
					webrcon.logger.Error("Failed to increment death count: ", err)
				}

				/*// Find the killer
				kid, err := db.Get("users", killerID)
				if err != nil {
					webrcon.logger.Trace("ERROR: ", err)
				}

				// TODO: Update the kill count for the killer
				if _, err := db.Set("users", map[string]interface{}{
					"SteamID": killerID,
					"Name":    killer}); err != nil {
					webrcon.logger.Trace("ERROR: ", err)
				}

				// TODO: Update the death count for the victim
				if _, err := db.Set("users", map[string]interface{}{
					"SteamID": victimID,
					"Name":    victim}); err != nil {
					webrcon.logger.Trace("ERROR: ", err)
					}*/
			}

			// TODO: I wonder if we should also send this to the game? Same for player join/leave?
			// Send the death message
			// webrcon.logger.Trace("Sending death message to Discord:", deathMessage)
			messageType := eventhandler.OtherKillType
			if isPvPKill {
				messageType = eventhandler.PvPKillType
			}
			webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: deathMessage, Type: messageType})
		} else {
			// webrcon.logger.Trace("Did not match any regex")
		}
	}
}

func (webrcon *Webrcon) handleIncomingDiscordMessage(message eventhandler.Message) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	webrcon.logger.Trace("handleIncomingDiscordMessage:", message)

	// Convert the message to a packet
	packet := Packet{Message: "say [DISCORD] " + message.User + ": " + message.Message, Identifier: 0, Type: "", Stacktrace: ""}

	// Convert the packet to a JSON string
	if jsonBytes, marshalErr := json.Marshal(packet); marshalErr != nil {
		webrcon.logger.Error("Failed to marshal webrcon packet:", marshalErr)
	} else {
		// Relay message to Webrcon
		webrcon.logger.Trace("Sending webrcon packet to server", string(jsonBytes))
		webrcon.Client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		webrcon.writeMutex.Lock()
		webrcon.Client.SendText(string(jsonBytes))
		webrcon.writeMutex.Unlock()
	}
}
