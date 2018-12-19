package discord

import (
	"log"
	"os"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/bwmarrin/discordgo"
)

// Client is the Discord client
var Client *discordgo.Session

var eventHandler *eventhandler.EventHandler
var webrconMessageHandler chan eventhandler.Message

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
	if _, channelSendMessageErr := Client.ChannelMessageSend(os.Getenv("DISCORD_BOT_CHANNEL_ID"), channelMessage); channelSendMessageErr != nil {
		log.Println("ERROR: Failed to send message to Discord:", channelSendMessageErr)
	}
}
