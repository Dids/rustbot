package webrcon

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/sacOO7/gowebsocket"
)

var eventHandler *eventhandler.EventHandler
var discordMessageHandler chan eventhandler.Message

var websocketClient gowebsocket.Socket

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
	}

	websocketClient.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		// log.Println("Received message " + message)

		// TODO: Also note that "Type: Chat" when it's a chat message!
		// FIXME: If "Identifier: -1", then we need to parse the packet a bit differently (NOTE: Looks like -1 is always a "chat packet"?):
		/*
			2018/12/19 10:00:25 Received message {
				"Message": "[CHAT] [ Kela ] Kotka[9574375/76561198402727161] : vähä aja päästä ois kai wipee jos kukaa on selannu tota discordia :D",
				"Identifier": 0,
				"Type": "Generic",
				"Stacktrace": ""
			}
			2018/12/19 10:00:25 Received message {
				"Message": "{\n  \"Message\": \"vähä aja päästä ois kai wipee jos kukaa on selannu tota discordia :D\",\n  \"UserId\": 76561198402727161,\n  \"Username\": \"[ Kela ] Kotka\",\n  \"Color\": \"#5af\",\n  \"Time\": 1545206425\n}",
				"Identifier": -1,
				"Type": "Chat",
				"Stacktrace": null
			}
		*/

		// Parse the incoming message as a webrcon packet
		packet := Packet{}
		if parseErr := json.Unmarshal([]byte(message), &packet); parseErr != nil {
			log.Println("ERROR: Failed to parse incoming message:", message, parseErr)
		}
		log.Println("Parsed message as packet:", packet)

		// TODO: Ideally we'd also need to parse connect and disconnect messages
		/*
			2018/12/19 11:09:10 Received message {
				"Message": "samppe[8194483/76561197991428801] has entered the game",
				"Identifier": 0,
				"Type": "Generic",
				"Stacktrace": ""
			}
			2018/12/19 10:56:25 Received message {
				"Message": "Saved 34,590 ents, cache(0.01), write(0.01), disk(0.00).",
				"Identifier": 0,
				"Type": "Generic",
				"Stacktrace": ""
			}
		*/

		// Handle different type conversions
		if packet.Identifier == ChatIdentifier && packet.Type == ChatType {
			chatPacket := ChatPacket{}
			if parseErr := json.Unmarshal([]byte(packet.Message), &chatPacket); parseErr != nil {
				log.Println("ERROR: Failed to parse incoming message:", parseErr)
			}
			log.Println("Parsed message as chat packet:", chatPacket)

			// Ignore messages from "SERVER"
			if chatPacket.Username == "SERVER" {
				log.Println("NOTICE: Ignoring message from SERVER")
				return
			}

			// Send chat message to Discord
			eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: chatPacket.Username, Message: chatPacket.Message})
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
	packet := Packet{Message: "say | " + message.User + ": " + message.Message, Identifier: 0, Type: "", Stacktrace: ""}

	// Convert the packet to a JSON string
	if jsonBytes, marshalErr := json.Marshal(packet); marshalErr != nil {
		log.Println("ERROR: Failed to marshal packet:", marshalErr)
	} else {
		// Relay message to Webrcon
		websocketClient.SendText(string(jsonBytes))
	}
}
