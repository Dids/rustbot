package discord

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Dids/rustbot/database"
	"github.com/Dids/rustbot/eventhandler"
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
	HasPresence           bool
	IsReady               bool
}

// NewDiscord creates and returns a new instance of Discord
func NewDiscord(handler *eventhandler.EventHandler, db *database.Database) (*Discord, error) {
	discord := &Discord{}

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

	// Setup our custom event handler
	discord.WebrconMessageHandler = make(chan eventhandler.Message)
	discord.EventHandler = handler
	discord.EventHandler.AddListener("receive_webrcon_message", discord.WebrconMessageHandler)
	go func() {
		for {
			discord.handleIncomingWebrconMessage(<-discord.WebrconMessageHandler)
		}
	}()

	// Store the database reference
	discord.Database = db

	return discord, nil
}

// Open will start the Discord client and connect to the API
func (discord *Discord) Open() error {
	return discord.Client.Open()
}

// Close will gracefully shutdown and cleanup the Discord client
func (discord *Discord) Close() error {
	discord.EventHandler.RemoveListener("receive_webrcon_message", discord.WebrconMessageHandler)
	return discord.Client.Close()
}

func (discord *Discord) handleConnect(session *discordgo.Session, event *discordgo.Connect) {
	log.Println("NOTICE: Discord event: connect")
	discord.IsReady = true
}

func (discord *Discord) handleDisconnect(session *discordgo.Session, event *discordgo.Disconnect) {
	log.Println("NOTICE: Discord event: disconnect")
	discord.IsReady = false
}

func (discord *Discord) handleRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	log.Println("NOTICE: Discord event: ratelimit")
}

func (discord *Discord) handleReady(session *discordgo.Session, event *discordgo.Ready) {
	log.Println("NOTICE: Discord event: ready")
	discord.IsReady = true
}

func (discord *Discord) handleMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
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

	// Only process messages from specific channels
	if channel.ID != os.Getenv("DISCORD_BOT_CHANNEL_ID") {
		log.Println("NOTICE: Ignoring message from channel:", "#"+channel.Name)
		return
	}

	// Relay the message to our message handler, which will eventually send it to the Webrcon client
	discord.EventHandler.Emit(eventhandler.Message{Event: "receive_discord_message", User: message.Author.Username, Message: message.Content})
}

func (discord *Discord) handleIncomingWebrconMessage(message eventhandler.Message) {
	// log.Println("handleIncomingWebrconMessage:", message)

	// Format any potential mentions
	mentionRegexMatches := mentionRegex.FindAllStringSubmatch(message.Message, -1)
	if len(mentionRegexMatches) > 0 {
		// Get the bot channel
		if botChannel, botChannelErr := discord.Client.Channel(os.Getenv("DISCORD_BOT_CHANNEL_ID")); botChannelErr != nil {
			log.Println("NOTICE: Failed to find bot channel:", botChannelErr)
		} else {
			// Get the bot guild from the channel
			if botGuild, botGuildErr := discord.Client.Guild(botChannel.GuildID); botGuildErr != nil {
				log.Println("NOTICE: Failed to find bot guild:", botGuildErr)
			} else {
				for _, match := range mentionRegexMatches {

					mentionUsername := strings.ToUpper(match[1])

					// Loop through each user in the guild
					for _, member := range botGuild.Members {
						username := strings.ToUpper(member.Nick)
						if username == "" {
							username = strings.ToUpper(member.User.Username)
						}

						if username == mentionUsername {
							replacer := newCaseInsensitiveReplacer(`@`+username, `<@!`+member.User.ID+`>`)
							formattedMentionMessage := replacer.Replace(message.Message)
							// log.Println("Formatted mention message:", formattedMentionMessage)
							message.Message = formattedMentionMessage

							break
						}
					}
				}
			}
		}
	}

	// TODO: Also replace the word "ylläpitäjä", "admin" and "admini" with "@Dids", so I get pinged? This should be configurable though..

	// Escape both "message.User" and "message.Message" to combat potential Markdown abuse
	//log.Println("Escaping message:", message)
	message = escapeMessage(message)
	//log.Println("Escaped message:", message)

	// Handle status messages
	if message.Type == eventhandler.StatusType {
		// Update presence
		// log.Println("Received status message, updating presence:", message.Message)
		if err := discord.updateNickname(message.User); err != nil {
			log.Println("NOTICE:", err)
		}
		if err := discord.updatePresence(message.Message); err != nil {
			log.Println("NOTICE:", err)
		}
		return
		// Handle server connect/disconnect messages
	} else if message.Type == eventhandler.ServerConnectedType || message.Type == eventhandler.ServerDisconnectedType {
		if _, err := discord.Client.ChannelMessageSend(os.Getenv("DISCORD_BOT_CHANNEL_ID"), "`"+message.Message+"`"); err != nil {
			log.Println("ERROR: Failed to send message to Discord:", message, err)
		}
		return
	} else if message.Type == eventhandler.PvPKillType || message.Type == eventhandler.OtherKillType {
		// Ignore PvP deaths if disabled
		if os.Getenv("KILLFEED_PVP_ENABLED") != "true" && message.Type == eventhandler.PvPKillType {
			// log.Println("Ignoring PvP kill, feed is disabled", os.Getenv("KILLFEED_PVP_ENABLED"))
			return
		}

		// Ignore Other deaths if disabled
		if os.Getenv("KILLFEED_OTHER_ENABLED") != "true" && message.Type == eventhandler.OtherKillType {
			// log.Println("Ignoring other kill, feed is disabled", os.Getenv("KILLFEED_OTHER_ENABLED"))
			return
		}

		// Send deaths to the main channel by default
		channelID := os.Getenv("DISCORD_BOT_CHANNEL_ID")

		// If the "kill feed" channel is set, override the channel
		if len(os.Getenv("KILLFEED_CHANNEL_ID")) > 0 {
			channelID = os.Getenv("KILLFEED_CHANNEL_ID")
		}

		if _, err := discord.Client.ChannelMessageSend(channelID, "_"+message.Message+"_"); err != nil {
			log.Println("ERROR: Failed to send message to Discord:", message, err)
		}

		return
	}

	// Format the message and send it to the specified channel
	channelMessage := "" + message.User + ": " + message.Message + ""
	if message.Type == eventhandler.JoinType || message.Type == eventhandler.DisconnectType {
		channelMessage = "_" + message.User + " " + string(message.Message) + "_"
	}
	if _, channelSendMessageErr := discord.Client.ChannelMessageSend(os.Getenv("DISCORD_BOT_CHANNEL_ID"), channelMessage); channelSendMessageErr != nil {
		log.Println("ERROR: Failed to send message to Discord:", message, channelSendMessageErr)
	}
}

func (discord *Discord) updateNickname(nickname string) error {
	if !discord.IsReady {
		return errors.New("Can't update nickname, Discord not ready")
	}

	// Set the nickname
	if discord.Client != nil && discord.Client.DataReady && nickname != "" {
		// Get the bot channel
		botChannel, botChannelErr := discord.Client.Channel(os.Getenv("DISCORD_BOT_CHANNEL_ID"))
		if botChannelErr != nil {
			return botChannelErr
		}

		// Construct the nickname payload
		data := struct {
			Nick string `json:"nick"`
		}{nickname}

		// Attempt to change the nickname using the Discord API
		_, updateNicknameErr := discord.Client.RequestWithBucketID("PATCH", discordgo.EndpointGuildMember(botChannel.GuildID, "@me")+"/nick", data, discordgo.EndpointGuildMember(botChannel.GuildID, ""))
		if updateNicknameErr != nil {
			return updateNicknameErr
		}
		// log.Println("Successfully updated the nickname:", string(updateNicknameResponse))
	} else {
		return errors.New("Can't update presence, Discord client is nil or not ready")
	}

	return nil
}

func (discord *Discord) updatePresence(presence string) error {
	if !discord.IsReady {
		return errors.New("Can't update presence, Discord not ready")
	}

	// Set the presence
	if discord.Client != nil && discord.Client.DataReady && presence != "" {
		if statusErr := discord.Client.UpdateStatus(0, presence); statusErr != nil {
			discord.HasPresence = false
			return statusErr
		}
		discord.HasPresence = true
	} else {
		return errors.New("Can't update presence, Discord client is nil or not ready")
	}

	return nil
}

// TODO: Refactor/move these to EventHandler, and if possible escape automatically!
func escapeMessage(message eventhandler.Message) eventhandler.Message {
	message.User = escapeMarkdown(message.User)
	message.Message = escapeMarkdown(message.Message)
	return message
}

func escapeMarkdown(markdown string) string {
	unescaped := unescapeBackslashRegex.ReplaceAllString(markdown, `$1`) // unescape any "backslashed" character
	escaped := escapeMarkdownRegex.ReplaceAllString(unescaped, `\$1`)
	return escaped
}

type caseInsensitiveReplacer struct {
	toReplace   *regexp.Regexp
	replaceWith string
}

func newCaseInsensitiveReplacer(toReplace, replaceWith string) *caseInsensitiveReplacer {
	return &caseInsensitiveReplacer{
		toReplace:   regexp.MustCompile("(?i)" + toReplace),
		replaceWith: replaceWith,
	}
}

func (cir *caseInsensitiveReplacer) Replace(str string) string {
	return cir.toReplace.ReplaceAllString(str, cir.replaceWith)
}
