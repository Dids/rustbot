package discord

import (
	"os"
	"regexp"

	"github.com/Dids/rustbot/database"
	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/logger"
	"github.com/bwmarrin/discordgo"
)

// Various precompiled regexes for Discord/Markdown
var mentionRegex = regexp.MustCompile(`(?:@)(\w+)`)
var unescapeBackslashRegex = regexp.MustCompile(`\\(\*|_|` + "`" + `|~|\\)`)
var escapeMarkdownRegex = regexp.MustCompile(`(\*|_|` + "`" + `|~|\\)`)

// Discord is an abstraction around the Discord client
type Discord struct {
	Client                *discordgo.Session
	EventHandler          *eventhandler.EventHandler
	Database              *database.Database
	WebrconMessageHandler chan eventhandler.Message
	LoggerMessageHandler  chan eventhandler.Message
	HasPresence           bool
	IsReady               bool

	// Private properties
	logger *logger.Logger
}

// NewDiscord creates and returns a new instance of Discord
func NewDiscord(handler *eventhandler.EventHandler, db *database.Database) (*Discord, error) {
	discord := &Discord{}

	// Store a reference to the Logger
	discord.logger = logger.GetLogger()

	// Initialize the Discord client
	if discordClient, discordClientErr := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN")); discordClientErr == nil {
		discord.Client = discordClient
	} else {
		return nil, discordClientErr
	}

	// Setup Discord client event handlers
	discord.Client.AddHandler(discord.handleConnect)
	discord.Client.AddHandler(discord.handleDisconnect)
	discord.Client.AddHandler(discord.handleRateLimit)
	discord.Client.AddHandler(discord.handleReady)
	discord.Client.AddHandler(discord.handleMessageCreate)

	// Setup our custom event handlers
	discord.WebrconMessageHandler = make(chan eventhandler.Message)
	discord.LoggerMessageHandler = make(chan eventhandler.Message)
	discord.EventHandler = handler
	discord.EventHandler.AddListener("receive_webrcon_message", discord.WebrconMessageHandler)
	discord.EventHandler.AddListener("receive_logger_message", discord.LoggerMessageHandler)
	go func() {
		for {
			// TODO: How does this actually work? Also do we need to stop this at some point? Do we need to use "select" etc. here?
			discord.handleIncomingWebrconMessage(<-discord.WebrconMessageHandler)
			discord.handleIncomingLoggerMessage(<-discord.LoggerMessageHandler) // TODO: Is this correct? This just runs in a loop and checks both, I guess?
		}
	}()

	// Store the database reference
	discord.Database = db

	return discord, nil
}

// Open will start the Discord client and connect to the API
func (discord *Discord) Open() error {
	discord.logger.Info("Opening Discord..")
	return discord.Client.Open()
}

// Close will gracefully shutdown and cleanup the Discord client
func (discord *Discord) Close() error {
	discord.logger.Info("Closing Discord..")
	discord.EventHandler.RemoveListener("receive_webrcon_message", discord.WebrconMessageHandler)
	return discord.Client.Close()
}
