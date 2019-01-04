package webrcon

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/sacOO7/gowebsocket"
)

var eventHandler *eventhandler.EventHandler
var discordMessageHandler chan eventhandler.Message
var websocketClient gowebsocket.Socket

var chatRegex = regexp.MustCompile(`\[CHAT\] (.+?)\[[0-9]+\/([0-9]+)\] : (.*)`)
var joinRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) joined \[(.*)\/([0-9]+)]`)
var disconnectRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) disconnecting: (.*)`)
var killRegex = regexp.MustCompile(`(?P<victim>.+?)(?:\[(?:[0-9]+?)\/(?P<victimid>[0-9]+?)\])(?: (?P<how>was killed by|died) )(?P<killer>(?:(?:[^\/\[\]]+)\[[0-9]+/(?P<killerid>[0-9]+)\]$)|(?P<reason>[^\/]*$))`)
var statusRegex = regexp.MustCompile(`(?:.*?hostname:\s*(?P<hostname>.*?)\\n)(?:.*?version\s*:\s*(?P<version>\d+) )(?:.*?secure\s*\((?P<secure>.*?)\)\\n)(?:.*?map\s*:\s*(?P<map>.*?)\\n)(?:.*?players\s*:\s*(?P<players_current>\d+) \((?P<players_max>\d+) max\) \((?P<players_queued>\d+) queued\) \((?P<players_joining>\d+) joining\)\\n)`)
var removeIDsRegex = regexp.MustCompile(`\[.+?\/.+?\]`)

// Status represents the current status of the server
var Status StatusPacket

// PacketType represents the type of a webrcon packet
type PacketType string

const (
	// GenericType is a packet type
	GenericType PacketType = "Generic"
	// ChatType is a packet type
	ChatType PacketType = "Chat"
	// IgnoreType is a packet type
	IgnoreType PacketType = "Ignore"
)

// PacketIdentifier represents the identifier of a webrcon packet
type PacketIdentifier int

const (
	// GenericIdentifier marks packets as generic data
	GenericIdentifier PacketIdentifier = 0
	// ChatIdentifier marks packets as chat messages
	ChatIdentifier PacketIdentifier = -1
	// IgnoreIdentifier marks packets as ignored
	IgnoreIdentifier PacketIdentifier = -2
)

// Packet represents a single webrcon packet
type Packet struct {
	Message    string           `json:"Message"`
	Identifier PacketIdentifier `json:"Identifier"`
	Type       PacketType       `json:"Type"`
	Stacktrace string           `json:"Stacktrace"`
}

// ChatPacket represents a single webrcon chat packet
type ChatPacket struct {
	Message  string `json:"Message"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
	Color    string `json:"Color"`
	Time     uint64 `json:"Time"`
}

// JoinPacket represents a single webrcon join packet
type JoinPacket struct {
	IP       string `json:"IP"`
	Port     string `json:"Port"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
	OS       string `json:"OS"`
}

// DisconnectPacket represents a single webrcon disconnect packet
type DisconnectPacket struct {
	IP       string `json:"IP"`
	Port     string `json:"Port"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
}

// StatusPacket represents a single webrcon status packet
type StatusPacket struct {
	Hostname       string `json:"hostname"`
	Version        int    `json:"version"`
	Secure         string `json:"secure"`
	Map            string `json:"map"`
	CurrentPlayers int    `json:"players_current"`
	MaxPlayers     int    `json:"players_max"`
	QueuedPlayers  int    `json:"players_queued"`
	JoiningPlayers int    `json:"players_joining"`
}

// Initialize will create and open a new Webrcon connection
func Initialize(handler *eventhandler.EventHandler) {
	log.Println("Initializing the Webrcon client..")

	// Initialize the websocket client/connection
	websocketClient = gowebsocket.New("ws://" + os.Getenv("WEBRCON_HOST") + ":" + os.Getenv("WEBRCON_PORT") + "/" + os.Getenv("WEBRCON_PASSWORD"))

	// Setup websocket event handlers
	websocketClient.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")

		// Send server connected message to Discord
		eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Connected to server!", Type: eventhandler.ServerConnectedType})
	}

	websocketClient.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server:", err)

		// Send server disconnected message to Discord
		eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Disconnected from server!", Type: eventhandler.ServerDisconnectedType})
	}

	websocketClient.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Received connect error ", err)

		// Send server disconnected message to Discord
		eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Cannot connect to server!", Type: eventhandler.ServerDisconnectedType})

		// Notify the primary process to shut down
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}

	websocketClient.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		//log.Println("Received message " + message)

		// FIXME: Remove these when done
		// message = `{ "Message": "109.240.100.173:18521/76561198806240991/Veru joined [windows/76561198806240991]", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		// message = `{ "Message": "109.240.100.173:18521/76561198806240991/Veru disconnecting: disconnect", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		// message = `{ "Message": "MurmeliOP[263066/76561198113377601] was killed by Vildemare[937684/76561198012399365]", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		message = `{ "Message": "Sarttuu[731399/76561198089400492] was killed by 7645878[29630/7645878]", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		// message = `{ "Message": "๖ۣۜZeUz[902806/76561197985407799] was killed by Hunger", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		// message = `{ "Message": "Tepachu[527565/76561198079774759] died (Fall)", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`

		// Parse the incoming message as a webrcon packet
		packet := Packet{}
		if parseErr := json.Unmarshal([]byte(message), &packet); parseErr != nil {
			log.Println("ERROR: Failed to parse as generic message:", message, parseErr)
		}
		//log.Println("Parsed message as packet:", packet)

		// TODO: These should be enabled/disabled through config/env vars
		// TODO: Handle events?
		//       {[event] assets/prefabs/npc/cargo plane/cargo_plane.prefab 0 Generic }

		// TODO: These should be enabled/disabled through config/env vars
		// TODO: Handle kill messages/packets:
		//       {MurmeliOP[263066/76561198113377601] was killed by Vildemare[937684/76561198012399365] 0 Generic }
		//       {Sarttuu[731399/76561198089400492] was killed by 7645878[29630/7645878] 0 Generic } // FIXME: According to @Karsiss, this is actually a scientist!
		//       {๖ۣۜZeUz[902806/76561197985407799] was killed by Hunger 0 Generic }

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
					// log.Println("Result (before):", string(result))
					result = statusRegex.Expand(result, statusTemplate, content, submatches)
					// log.Println("Result (after):", string(result))
				}
				// log.Println("End result:", string(result))

				// Convert the resulting JSON string to a StatusPacket
				if err := json.Unmarshal(result, &Status); err != nil {
					log.Println("ERROR: Failed to parse status message:", err)
				} else {
					// log.Println("Received new status:", Status)

					// TODO: Refactor the format like so:
					//       Playing "1/64 (2 joining, 3 queued)"
					//       Playing "1/64 (2 joining)"
					//       Playing "1/64 (2 queued)"
					//       Playing "1/64"

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
					// log.Println("Status updated, emitting status message:", message)
					eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: Status.Hostname, Message: message, Type: eventhandler.StatusType})
					return
				}
			}
		}

		//log.Println("Parsed message as packet:", packet)

		// Handle different type conversions
		if packet.Identifier == ChatIdentifier && packet.Type == ChatType {
			chatPacket := ChatPacket{}
			if parseErr := json.Unmarshal([]byte(packet.Message), &chatPacket); parseErr != nil {
				log.Println("ERROR: Failed to parse as chat message:", message, parseErr)
			}
			// log.Println("Parsed message as chat packet:", chatPacket)

			// Ignore messages from "SERVER"
			if chatPacket.Username == "SERVER" {
				// log.Println("NOTICE: Ignoring message from SERVER")
				return
			}

			// Send chat message to Discord
			eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: chatPacket.Username, Message: chatPacket.Message})
		} else {
			joinRegexMatches := joinRegex.FindStringSubmatch(packet.Message)
			disconnectRegexMatches := disconnectRegex.FindStringSubmatch(packet.Message)
			killRegexMatches := killRegex.FindStringSubmatch(packet.Message)
			if len(joinRegexMatches) > 1 {
				// log.Println("Matched joinRegex:", joinRegexMatches)
				userID, _ := strconv.ParseUint(joinRegexMatches[3], 10, 64)
				joinPacket := JoinPacket{IP: joinRegexMatches[1], Port: joinRegexMatches[2], UserID: userID, Username: joinRegexMatches[4], OS: joinRegexMatches[5]}
				// log.Println("Join packet:", joinPacket)
				eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: joinPacket.Username, Message: "joined", Type: eventhandler.JoinType})
			} else if len(disconnectRegexMatches) > 1 {
				// log.Println("Matched disconnectRegex:", disconnectRegexMatches)
				userID, _ := strconv.ParseUint(disconnectRegexMatches[3], 10, 64)
				disconnectPacket := DisconnectPacket{IP: disconnectRegexMatches[1], Port: disconnectRegexMatches[2], UserID: userID, Username: disconnectRegexMatches[4]}
				// log.Println("Disconnect packet:", disconnectPacket)
				eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: disconnectPacket.Username, Message: "left", Type: eventhandler.DisconnectType})
			} else if len(killRegexMatches) > 1 {
				// Construct a simple "dictionary" using the named capture groups
				result := make(map[string]string)
				for i, name := range killRegex.SubexpNames() {
					if i != 0 && name != "" {
						result[name] = killRegexMatches[i]
					}
				}

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

				// Reformat the death reason
				reason = strings.ToLower(reason)
				reason = strings.Replace(reason, "(", "", -1)
				reason = strings.Replace(reason, ")", "", -1)

				/*log.Println("killRegexMatches:", killRegexMatches)
				log.Println("result:", result)
				log.Println("victim:", victim)
				log.Println("victimID:", victimID)
				log.Println("how:", how)
				log.Println("killer:", killer)
				log.Println("killerID:", killerID)
				log.Println("reason:", reason)*/

				// TODO: These still needs work, for situations like this:
				//       6213551 died from generic

				// Skip scientist deaths
				if len(victim) > 0 && len(victimID) > 0 && victim == victimID {
					//log.Println("NOTICE: Skippig potential scientist death:", message)
					return
				}

				// Rename scientists
				if len(killer) > 0 && len(killerID) > 0 && killer == killerID {
					killer = "a scientist"
				}

				// Construct the death message
				deathMessage := ""
				if len(victim) > 0 && len(how) > 0 && len(killer) > 0 && len(reason) == 0 {
					// "PlayerA was killed by PlayerB"
					deathMessage = victim + " " + how + " " + killer
				} else if len(victim) > 0 && len(how) > 0 && len(reason) > 0 {
					if len(how) == 4 {
						// "PlayerA died fall"
						deathMessage = victim + " " + how + " from " + reason
					} else {
						// "PlayerA was killed by Hunger"
						deathMessage = victim + " " + how + " " + reason
					}
				} else {
					// TODO: What if our error handler DM'd us any errors? That'd be super cool and useful!
					log.Println("NOTICE: Could not parse death message:", message)
					return
				}

				// TODO: I wonder if we should also send this to the game? Same for player join/leave?
				// Send the death message
				//log.Println("Sending death message to Discord:", deathMessage)
				eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: deathMessage, Type: eventhandler.KillType})
			} else {
				// log.Println("Did not match any regex")
			}
		}
	}

	websocketClient.OnBinaryMessage = func(data []byte, socket gowebsocket.Socket) {
		log.Println("Received binary data ", data)
	}

	websocketClient.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		// log.Println("Received ping " + data)
	}

	websocketClient.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong " + data)
	}

	// FIXME: We need to be able to reconnect, or at the very least exit the process, so we can at least recover that way
	websocketClient.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")

		// Notify the primary process to shut down
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}

	// Setup our custom event handler
	discordMessageHandler = make(chan eventhandler.Message)
	eventHandler = handler
	eventHandler.AddListener("receive_discord_message", discordMessageHandler)
	go func() {
		for {
			handleIncomingDiscordMessage(<-discordMessageHandler)
		}
	}()

	// Finally establish the websocket connetion
	websocketClient.Connect()

	// Start updating status
	go startUpdatingStatus()

	log.Println("Successfully created the Webrcon client!")
}

// Close will gracefully shutdown and cleanup the Webrcon connection
func Close() {
	// Send shutdown message to Discord
	eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Going away, see you in a bit!", Type: eventhandler.ServerDisconnectedType})

	// Sleep for a bit before shutting down
	time.Sleep(1 * time.Second)

	log.Println("Shutting down the Webrcon client..")
	eventHandler.RemoveListener("receive_discord_message", discordMessageHandler)
	websocketClient.Close()
	log.Println("Successfully shut down the Webrcon client!")
}

func handleIncomingDiscordMessage(message eventhandler.Message) {
	log.Println("handleIncomingDiscordMessage:", message)

	// Convert the message to a packet
	packet := Packet{Message: "say [DISCORD] " + message.User + ": " + message.Message, Identifier: 0, Type: "", Stacktrace: ""}

	// Convert the packet to a JSON string
	if jsonBytes, marshalErr := json.Marshal(packet); marshalErr != nil {
		log.Println("ERROR: Failed to marshal packet:", marshalErr)
	} else {
		// Relay message to Webrcon
		// log.Println("!!! SENDING DATA TO WEBRCON SERVER !!!", string(jsonBytes))
		websocketClient.SendText(string(jsonBytes))
	}
}

// TODO: Rewrite this at some point so we can stop it if we want to,
//       for example if we start updating after the "ready" websocket
//       event, as that's where it would make the most sense, right?
func startUpdatingStatus() {
	for {
		// Send "status" command
		// log.Println("Requesting status..")
		websocketClient.SendText(`{ "Message": "status", "Identifier": 0, "Type": "Generic" }`)

		// Sleep for a bit before requesting the status (Discord API only allows the presence to be updated every 15 seconds)
		time.Sleep(15 * time.Second)
	}
}
