package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dids/rustbot/discord"
	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/webrcon"
	"github.com/joho/godotenv"
)

var eventHandler eventhandler.EventHandler

func main() {
	// Print a banner
	fmt.Println("-----------------")
	fmt.Println("---  RustBot  ---")
	fmt.Println("-----------------")
	fmt.Println("")

	// Load environment variables
	godotenvErr := godotenv.Load()
	if godotenvErr != nil {
		log.Fatal("Error loading .env file:", godotenvErr)
	}

	// Initialize our own event handler
	eventHandler = eventhandler.EventHandler{Name: "rustbot", Listeners: nil}

	// Initialize the Discord client
	discord.Initialize(&eventHandler)

	// Initialize the Webrcon Client
	webrcon.Initialize(&eventHandler)

	// TODO: Implement and setup event handlers for both Discord and Webrcon clients, so they can pass messages between each other

	// TODO: Wait for CTRL-C or something, then call <module>.close() when shutting down
	// Wait here until CTRL-C or other term signal is received.
	log.Println("RustBot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Properly dispose of the clients when exiting
	webrcon.Close()
	discord.Close()
}
