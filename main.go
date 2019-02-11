package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Dids/rustbot/database"
	"github.com/Dids/rustbot/discord"
	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/logger"
	"github.com/Dids/rustbot/webrcon"

	_ "github.com/joho/godotenv/autoload"
)

var eventHandler eventhandler.EventHandler

func main() {
	// Initialize our own event handler
	eventHandler = eventhandler.EventHandler{Name: "rustbot", Listeners: nil}

	// Determine our log level
	logLevel := logger.Trace
	if len(os.Getenv("LOG_LEVEL")) > 0 {
		envValue := os.Getenv("LOG_LEVEL")
		parsedValue, _ := strconv.Atoi(envValue)
		logLevel = logger.LogLevel(parsedValue)
	}

	// Initialize our logger
	loggerOptions := logger.Options{
		Level: logLevel,
		File:  "rustbot.log", // TODO: Test to make sure that this doesn't need any extra path handling!
	}
	logger := logger.NewLogger(loggerOptions, &eventHandler)

	// Print a message to signal that we're starting
	logger.Info("Starting..")

	// Initialize and open the Database
	database, databaseErr := database.NewDatabase()
	if databaseErr != nil {
		logger.Panic("Failed to initialize database:", databaseErr)
	}
	if databaseErr = database.Open(); databaseErr != nil {
		logger.Panic("Failed to open database:", databaseErr)
	}

	// Initialize and open the Discord client
	discord, discordErr := discord.NewDiscord(&eventHandler, database)
	if discordErr != nil {
		logger.Panic("Failed to initialize Discord:", discordErr)
	}
	if discordErr = discord.Open(); discordErr != nil {
		logger.Panic("Failed to open Discord:", discordErr)
	}

	// Initialize and open the Webrcon client
	webrcon, webrconErr := webrcon.NewWebrcon(&eventHandler, database)
	if webrconErr != nil {
		logger.Panic("Failed to initialize Webrcon:", webrconErr)
	}
	if webrconErr := webrcon.Open(); webrconErr != nil {
		logger.Panic("Failed to open Webrcon:", webrconErr)
	}

	// Wait here until CTRL-C or other term signal is received.
	logger.Info("RustBot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Print a message to signal that we're stopping
	logger.Info("Stopping..")

	// Properly dispose of the clients when exiting
	if err := webrcon.Close(); err != nil {
		logger.Panic("Failed to close Webrcon:", err)
	}
	if err := discord.Close(); err != nil {
		logger.Panic("Failed to close Discord:", err)
	}
	if err := database.Close(); err != nil {
		logger.Panic("Failed to close database:", err)
	}
}
