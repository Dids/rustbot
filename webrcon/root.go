package webrcon

import (
	"log"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/webrcon/client"
)

var eventHandler *eventhandler.EventHandler
var discordMessageHandler chan eventhandler.Message

// Initialize will create and open a new Webrcon connection
func Initialize(handler *eventhandler.EventHandler) {
	log.Println("Initializing the Webrcon client..")
	client.Initialize()
	log.Println("Successfully created the Webrcon client!")

	// Setup our custom event handler
	discordMessageHandler = make(chan eventhandler.Message)
	eventHandler = handler
	eventHandler.AddListener("receive_discord_message", discordMessageHandler)
	go func() {
		for {
			handleIncomingDiscordMessage(<-discordMessageHandler)
		}
	}()

	// TODO: !!! Make absolutely sure to ignore "RustBot" messages though, so we don't end up with recursive loops and such !!!
	// TODO: Setup Webrcon -> Discord relaying, like this:
	// eventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: user, Message: message})
}

// Close will gracefully shutdown and cleanup the Webrcon connection
func Close() {
	log.Println("Shutting down the Webrcon client..")
	eventHandler.RemoveListener("receive_discord_message", discordMessageHandler)
	client.Close()
	log.Println("Successfully shut down the Webrcon client!")
}

func handleIncomingDiscordMessage(message eventhandler.Message) {
	log.Println("handleIncomingDiscordMessage:", message)

	// TODO: Relay message to Webrcon
}
