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

	webrcon.logger.Trace("Received message " + message)

	// Parse the incoming message as a webrcon packet
	packet := Packet{}
	if parseErr := json.Unmarshal([]byte(message), &packet); parseErr != nil {
		webrcon.logger.Error("Failed to parse as generic message:", message, parseErr)
	}
	//webrcon.logger.Trace("Parsed message as packet:", packet)

	if packet.Identifier == GenericIdentifier && packet.Type == GenericType {
		// Check if this is a valid status message
		statusRegexMatches := statusRegex.FindStringSubmatch(message)
		if len(statusRegexMatches) > 0 {
			// Template for converting status message to a JSON string
			statusTemplate := []byte(`{ "hostname": "$hostname", "version": $version, "secure": "$secure", "map": "$map", "players_current": $players_current, "players_max": $players_max, "players_queued": $players_queued, "players_joining": $players_joining }`)
			result := []byte{}
			content := []byte(message)

			// TODO: Isn't this sort of redundant, since we'll never have more than 1 match anyway?
			// For each match of the regex in the content
			for _, submatches := range statusRegex.FindAllSubmatchIndex(content, -1) {
				// Apply the captured submatches to the template and append the output to the result
				// webrcon.logger.Trace("Result (before):", string(result))
				result = statusRegex.Expand(result, statusTemplate, content, submatches)
				// webrcon.logger.Trace("Result (after):", string(result))
			}
			// webrcon.logger.Trace("End result:", string(result))

			// Convert the resulting JSON string to a StatusPacket
			if err := json.Unmarshal(result, &Status); err != nil {
				webrcon.logger.Error("Failed to parse status message:", err)
			} else {
				// webrcon.logger.Trace("Received new status:", Status)

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
