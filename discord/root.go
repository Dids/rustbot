package discord

import (
	"log"
	"os"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/bwmarrin/discordgo"
)

// Client is the Discord client
var Client *discordgo.Session

var eventHandler *eventhandler.EventHandler
var webrconMessageHandler chan eventhandler.Message
var hasPresence bool

// Initialize will create and start a new Discord client
func Initialize(handler *eventhandler.EventHandler) {
	// Initialize the Discord client
	log.Println("Initializing the Discord client..")
	if discordClient, discordClientErr := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN")); discordClientErr == nil {
		Client = discordClient
	} else {
		log.Fatal("Error creating Discord client:", discordClientErr)
	}
	log.Println("Successfully created the Discord client")

	// Setup Discord client event handlers
	Client.AddHandler(handleIncomingMessage)

	// FIXME: Do we have a "ready handler" that we could use to set the presence?

	// Start updating presence
	go startUpdatingPresence()

	// Setup our custom event handler
	webrconMessageHandler = make(chan eventhandler.Message)
	eventHandler = handler
	eventHandler.AddListener("receive_webrcon_message", webrconMessageHandler)
	go func() {
		for {
			handleIncomingWebrconMessage(<-webrconMessageHandler)
		}
	}()

	// Open a websocket connection to Discord and begin listening.
	discordClientOpenErr := Client.Open()
	if discordClientOpenErr != nil {
		log.Fatal("Error opening Discord client connection:", discordClientOpenErr)
	}
}

// Close will gracefully shutdown and cleanup the Discord client
func Close() {
	log.Println("Shutting down the Discord client..")
	eventHandler.RemoveListener("receive_webrcon_message", webrconMessageHandler)
	stopUpdatingPresence()
	Client.Close()
	log.Println("Successfully shut down the Discord client!")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func handleIncomingMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}

	// Find the channel that the message originated from
	channel, channelErr := session.State.Channel(message.ChannelID)
	if channelErr != nil {
		log.Println("WARNING: Could not find channel with ID", message.ChannelID)
		return
	}

	// TODO: Eventually change this to "#rust" (or ideally expose this via a startup parameter)
	// Only process messages from specific channels
	if channel.ID != os.Getenv("DISCORD_BOT_CHANNEL_ID") {
		log.Println("NOTICE: Ignoring message from channel:", "#"+channel.Name)
		return
	}

	// Relay the message to our message handler, which will eventually send it to the Webrcon client
	eventHandler.Emit(eventhandler.Message{Event: "receive_discord_message", User: message.Author.Username, Message: message.Content})
}

func handleIncomingWebrconMessage(message eventhandler.Message) {
	log.Println("handleIncomingWebrconMessage:", message)

	// Format the message and send it to the specified channel
	channelMessage := "`" + message.User + ": " + message.Message + "`"
	if message.Type == eventhandler.JoinType || message.Type == eventhandler.DisconnectType {
		channelMessage = "`* " + message.User + " " + string(message.Message) + "`"
	}
	if _, channelSendMessageErr := Client.ChannelMessageSend(os.Getenv("DISCORD_BOT_CHANNEL_ID"), channelMessage); channelSendMessageErr != nil {
		log.Println("ERROR: Failed to send message to Discord:", channelSendMessageErr)
	}
}

func startUpdatingPresence() {
	for {
		// Sleep for a bit before updating the presence
		if hasPresence {
			time.Sleep(5 * time.Minute)
		} else {
			time.Sleep(15 * time.Second)
		}

		// Update presence
		updatePresence(os.Getenv("WEBRCON_HOST") + ":" + "28015")
	}
}

func updatePresence(presence string) error {
	// Set the presence
	if Client != nil && Client.DataReady && presence != "" {
		if statusErr := Client.UpdateStatus(0, presence); statusErr != nil {
			log.Println("NOTICE: Failed to update presence:", statusErr)
			hasPresence = false
			return statusErr
		} else {
			hasPresence = true
		}
	}
	return nil
}

func stopUpdatingPresence() {
	// Stop the timer
	// updatePresenceTimer.Stop()
}
