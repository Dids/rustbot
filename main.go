package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dids/rustbot/database"
	"github.com/Dids/rustbot/discord"
	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/webrcon"

	_ "github.com/joho/godotenv/autoload"
)

var eventHandler eventhandler.EventHandler

func main() {
	// Print a banner
	log.Println("-----------------")
	log.Println("---  RustBot  ---")
	log.Println("-----------------")
	log.Println("")

	// Initialize our own event handler
	eventHandler = eventhandler.EventHandler{Name: "rustbot", Listeners: nil}

	// Initialize and open the Database
	database, databaseErr := database.NewDatabase()
	if databaseErr != nil {
		log.Panic("Failed to initialize database:", databaseErr)
	}
	if databaseErr = database.Open(); databaseErr != nil {
		log.Panic("Failed to open database:", databaseErr)
	}

	// Initialize and open the Discord client
	discord, discordErr := discord.NewDiscord(&eventHandler, database)
	if discordErr != nil {
		log.Panic("Failed to initialize Discord:", discordErr)
	}
	if discordErr = discord.Open(); discordErr != nil {
		log.Panic("Failed to open Discord:", discordErr)
	}

	// TODO: Shouldn't we follow the same logic here, so having a separate "Open()" function?
	// Initialize the Webrcon Client (opens the connection automatically)
	webrcon.Initialize(&eventHandler, database)

	// TODO: Implement and setup event handlers for both Discord and Webrcon clients, so they can pass messages between each other

	// TODO: Wait for CTRL-C or something, then call <module>.close() when shutting down
	// Wait here until CTRL-C or other term signal is received.
	log.Println("RustBot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Properly dispose of the clients when exiting
	webrcon.Close()
	if err := discord.Close(); err != nil {
		log.Panic(err)
	}
	if err := database.Close(); err != nil {
		log.Panic(err)
	}
}
