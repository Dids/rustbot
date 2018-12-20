package webrcon

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/sacOO7/gowebsocket"
)

var eventHandler *eventhandler.EventHandler
var discordMessageHandler chan eventhandler.Message
var websocketClient gowebsocket.Socket

var chatRegex = regexp.MustCompile(`\[CHAT\] (.+?)\[[0-9]+\/([0-9]+)\] : (.*)`)
var joinRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) joined \[(.*)\/([0-9]+)]`)
var disconnectRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) disconnecting: (.*)`)

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

// Initialize will create and open a new Webrcon connection
func Initialize(handler *eventhandler.EventHandler) {
	log.Println("Initializing the Webrcon client..")

	// Initialize the websocket client/connection
	websocketClient = gowebsocket.New("ws://" + os.Getenv("WEBRCON_HOST") + ":" + os.Getenv("WEBRCON_PORT") + "/" + os.Getenv("WEBRCON_PASSWORD"))

	// Setup websocket event handlers
	websocketClient.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	websocketClient.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Received connect error ", err)

		// Notify the primary process to shut down
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}

	websocketClient.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		// log.Println("Received message " + message)

		// FIXME: Remove these when done
		// message = `{ "Message": "109.240.100.173:18521/76561198806240991/Veru joined [windows/76561198806240991]", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`
		// message = `{ "Message": "109.240.100.173:18521/76561198806240991/Veru disconnecting: disconnect", "Identifier": 0, "Type": "Generic", "StackTrace": "" }`

		// Parse the incoming message as a webrcon packet
		packet := Packet{}
		if parseErr := json.Unmarshal([]byte(message), &packet); parseErr != nil {
			log.Println("ERROR: Failed to parse as generic message:", message, parseErr)
		}
		// log.Println("Parsed message as packet:", packet)

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

	log.Println("Successfully created the Webrcon client!")
}

// Close will gracefully shutdown and cleanup the Webrcon connection
func Close() {
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
