package discord

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/Dids/rustbot/webrcon"
	"github.com/bwmarrin/discordgo"
)

func (discord *Discord) handleMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}

	// Find the channel that the message originated from
	channel, channelErr := session.State.Channel(message.ChannelID)
	if channelErr != nil {
		discord.logger.Warning("Could not find channel with ID", message.ChannelID)
		return
	}

	// Only process messages from specific channels
	if channel.ID != os.Getenv("DISCORD_CHAT_CHANNEL_ID") {
		discord.logger.Trace("Ignoring message from channel:", "#"+channel.Name)
		return
	}

	// Relay the message to our message handler, which will eventually send it to the Webrcon client
	discord.EventHandler.Emit(eventhandler.Message{Event: "receive_discord_message", User: message.Author.Username, Message: message.Content})
}

func (discord *Discord) handleIncomingLoggerMessage(message eventhandler.Message) {
	// TODO: Don't use discord.logger.* here, because that'd cause recursion, right?
	//log.Println("handleIncomingLoggerMessage:", message)
	//discord.logger.Trace("handleIncomingLoggerMessage:", message)

	// Escape the message
	message = escapeMessage(message)

	// Send the message to the logs channel as an embed
	if len(os.Getenv("DISCORD_LOG_CHANNEL_ID")) > 0 {
		messageTypeIcon := "❓"
		switch messageType := message.Type; messageType {
		case eventhandler.TraceLogType:
			messageTypeIcon = "🔧"
		case eventhandler.InfoLogType:
			messageTypeIcon = "ℹ"
		case eventhandler.WarningLogType:
			messageTypeIcon = "⚠"
		case eventhandler.ErrorLogType:
			messageTypeIcon = "❗"
		case eventhandler.PanicLogType:
			messageTypeIcon = "⛔"
		}

		embedFields := make([]*discordgo.MessageEmbedField, 1)
		embedFields[0] = &discordgo.MessageEmbedField{
			Name:   messageTypeIcon,
			Value:  "`" + string(message.Message) + "`",
			Inline: false,
		}

		embed := &discordgo.MessageEmbed{
			Timestamp: time.Now().Format(time.RFC3339),
			Fields:    embedFields,
		}

		if _, err := discord.Client.ChannelMessageSendEmbed(os.Getenv("DISCORD_LOG_CHANNEL_ID"), embed); err != nil {
			log.Println("NOTICE: Failed to send message to logs channel:", message, "with error:", err)
			//discord.logger.Error("Failed to send message to logs channel:", message, "with error:", err)
		}
	}

	// We also want to send panic logs to the admin as a Direct Message (only if the admin is set)
	if message.Type == eventhandler.PanicLogType && len(os.Getenv("DISCORD_OWNER_ID")) > 0 {
		ownerChannel, err := discord.Client.UserChannelCreate(os.Getenv("DISCORD_OWNER_ID"))
		if err != nil {
			log.Println("NOTICE: Failed to create admin DM for message:", message, "with error:", err)
			//discord.logger.Error("Failed to send message to admin:", message, "with error:", err)
		} else {
			if _, err := discord.Client.ChannelMessageSend(ownerChannel.ID, string(message.Message)); err != nil {
				log.Println("NOTICE; Failed to send message to admin:", message, "with error:", err)
				//discord.logger.Error("Failed to send message to admin:", message, "with error:", err)
			}
		}
	}
}

func (discord *Discord) handleIncomingWebrconMessage(message eventhandler.Message) {
	discord.logger.Trace("handleIncomingWebrconMessage:", message)

	// Format any potential mentions
	mentionRegexMatches := mentionRegex.FindAllStringSubmatch(message.Message, -1)
	if len(mentionRegexMatches) > 0 {
		// Get the bot channel
		if botChannel, botChannelErr := discord.Client.Channel(os.Getenv("DISCORD_CHAT_CHANNEL_ID")); botChannelErr != nil {
			discord.logger.Warning("Failed to find bot channel:", botChannelErr)
		} else {
			// Get the bot guild from the channel
			if botGuild, botGuildErr := discord.Client.Guild(botChannel.GuildID); botGuildErr != nil {
				discord.logger.Warning("Failed to find bot guild:", botGuildErr)
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
							// discord.logger.Trace("Formatted mention message:", formattedMentionMessage)
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
	//discord.logger.Trace("Escaping message:", message)
	message = escapeMessage(message)
	//discord.logger.Trace("Escaped message:", message)

	// Handle status messages
	if message.Type == eventhandler.StatusType {
		// Update presence
		discord.logger.Trace("Received status message, updating presence:", message.Message)
		if err := discord.updateNickname(message.User); err != nil {
			discord.logger.Error("Failed to update nickname:", err)
		}
		if err := discord.updatePresence(message.Message); err != nil {
			discord.logger.Error("Failed to update presence:", err)
		}
		return
		// Handle server connect/disconnect messages
	} else if message.Type == eventhandler.PlayersType {
		// Update player list
		discord.logger.Trace("Received players message, updating players:", message.Message)
		parsedPlayers := make([]webrcon.PlayerPacket, 0)
		if err := json.Unmarshal([]byte(message.Message), &parsedPlayers); err != nil {
			discord.logger.Error("Failed to parse players:", err)
		}
		if err := discord.updatePlayers(parsedPlayers); err != nil {
			discord.logger.Error("Failed to update players:", err)
		}
		return
	} else if message.Type == eventhandler.ServerConnectedType || message.Type == eventhandler.ServerDisconnectedType {
		if _, err := discord.Client.ChannelMessageSend(os.Getenv("DISCORD_CHAT_CHANNEL_ID"), "`"+message.Message+"`"); err != nil {
			discord.logger.Error("Failed to send message:", message, "with error:", err)
		}
		return
	} else if message.Type == eventhandler.PvPKillType || message.Type == eventhandler.OtherKillType {
		// Ignore PvP deaths if disabled
		if os.Getenv("DISCORD_KILLFEED_PVP_ENABLED") != "true" && message.Type == eventhandler.PvPKillType {
			discord.logger.Trace("Ignoring PvP kill, feed is disabled", os.Getenv("DISCORD_KILLFEED_PVP_ENABLED"))
			return
		}

		// Ignore Other deaths if disabled
		if os.Getenv("DISCORD_KILLFEED_OTHER_ENABLED") != "true" && message.Type == eventhandler.OtherKillType {
			discord.logger.Trace("Ignoring other kill, feed is disabled", os.Getenv("DISCORD_KILLFEED_OTHER_ENABLED"))
			return
		}

		// Send deaths to the main channel by default
		channelID := os.Getenv("DISCORD_CHAT_CHANNEL_ID")

		// If the "kill feed" channel is set, override the channel
		if len(os.Getenv("DISCORD_KILLFEED_CHANNEL_ID")) > 0 {
			channelID = os.Getenv("DISCORD_KILLFEED_CHANNEL_ID")
		}

		if _, err := discord.Client.ChannelMessageSend(channelID, "_"+message.Message+"_"); err != nil {
			discord.logger.Error("Failed to send message:", message, "with error:", err)
		}

		return
	} else if message.Type == eventhandler.JoinType || message.Type == eventhandler.DisconnectType {
		// Send join/leave messages to the main channel by default
		channelID := os.Getenv("DISCORD_CHAT_CHANNEL_ID")

		// If the "notifications" channel is set, override the channel
		if len(os.Getenv("DISCORD_NOTIFICATIONS_CHANNEL_ID")) > 0 {
			channelID = os.Getenv("DISCORD_NOTIFICATIONS_CHANNEL_ID")
		}

		if _, err := discord.Client.ChannelMessageSend(channelID, "_"+message.User+" "+string(message.Message)+"_"); err != nil {
			discord.logger.Error("Failed to send message:", message, "with error:", err)
		}

		return
	}

	// Skip the message of the user is missing
	if len(message.User) <= 0 {
		discord.logger.Warning("Skipping chat message with missing username:", message)
		return
	}

	// Format the message and send it to the specified channel
	channelMessage := "" + message.User + ": " + message.Message + ""
	if message.Type == eventhandler.JoinType || message.Type == eventhandler.DisconnectType {
		channelMessage = "_" + message.User + " " + string(message.Message) + "_"
	}
	if _, err := discord.Client.ChannelMessageSend(os.Getenv("DISCORD_CHAT_CHANNEL_ID"), channelMessage); err != nil {
		discord.logger.Error("Failed to send message:", message, "with error:", err)
	}
}
