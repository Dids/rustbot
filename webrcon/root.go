package webrcon

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/Dids/rustbot/database"
	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/logger"
	"github.com/HouzuoGuo/tiedot/db"

	"github.com/sacOO7/gowebsocket"
)

// Webrcon is an abstraction around the Webrcon client
type Webrcon struct {
	Client                gowebsocket.Socket
	EventHandler          *eventhandler.EventHandler
	Database              *database.Database
	DiscordMessageHandler chan eventhandler.Message

	// Private properties
	logger          *logger.Logger
	usersCollection *db.Col
	isShuttingDown  bool
	writeMutex      *sync.Mutex
}

// NewWebrcon creates and returns a new instance of Webrcon
func NewWebrcon(handler *eventhandler.EventHandler, db *database.Database) (*Webrcon, error) {
	webrcon := &Webrcon{}

	// Store a reference to the Logger
	webrcon.logger = logger.GetLogger()

	// Store a reference to the users collection
	result, err := db.GetCollection("users")
	if err != nil {
		return nil, err
	}
	webrcon.usersCollection = result

	// Make sure that required indexes are set on the users collection
	webrcon.usersCollection.Index([]string{"SteamID"})

	// Initialize the websocket client
	webrcon.Client = gowebsocket.New("ws://" + os.Getenv("WEBRCON_HOST") + ":" + os.Getenv("WEBRCON_PORT") + "/" + os.Getenv("WEBRCON_PASSWORD"))

	// Setup websocket client event handlers
	webrcon.Client.OnConnected = webrcon.handleConnect
	webrcon.Client.OnDisconnected = webrcon.handleDisconnect
	webrcon.Client.OnConnectError = webrcon.handleConnectError
	webrcon.Client.OnPingReceived = webrcon.handlePingReceived
	webrcon.Client.OnPongReceived = webrcon.handlePongReceived
	webrcon.Client.OnTextMessage = webrcon.handleTextMessage

	// Setup our custom event handler
	webrcon.DiscordMessageHandler = make(chan eventhandler.Message)
	webrcon.EventHandler = handler
	webrcon.EventHandler.AddListener("receive_discord_message", webrcon.DiscordMessageHandler)
	go func() {
		for {
			webrcon.handleIncomingDiscordMessage(<-webrcon.DiscordMessageHandler)
		}
	}()

	// Store the database reference
	webrcon.Database = db

	// Create the write mutex
	webrcon.writeMutex = &sync.Mutex{}

	return webrcon, nil
}

// Open will start the Webrcon client and connect to the server
func (webrcon *Webrcon) Open() error {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return errors.New("shutdown in progress")
	}

	webrcon.logger.Info("Opening Webrcon..")

	// Establish the websocket connetion
	webrcon.Client.Connect()

	// Start sending PING messages
	go webrcon.startPinging()

	// Start updating status
	go webrcon.startUpdatingStatus()

	return nil
}

// Close will gracefully shutdown and cleanup the Webrcon connection
func (webrcon *Webrcon) Close() error {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return errors.New("shutdown in progress")
	}
	webrcon.isShuttingDown = true

	webrcon.logger.Info("Closing Webrcon..")

	// Send shutdown message to Discord
	webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Going away, see you in a bit..", Type: eventhandler.ServerDisconnectedType})

	// Sleep for a bit before shutting down
	time.Sleep(1 * time.Second)

	webrcon.EventHandler.RemoveListener("receive_discord_message", webrcon.DiscordMessageHandler)
	webrcon.Client.Close()
	webrcon.logger.Trace("Successfully shut down the Webrcon client!")

	return nil
}
