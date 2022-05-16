package discord

import (
	"errors"
	"os"

	"github.com/bwmarrin/discordgo"
)

func (discord *Discord) updateNickname(nickname string) error {
	if !discord.IsReady {
		return errors.New("Can't update nickname, Discord not ready (discord.IsReady = false)")
	}

	// Set the nickname
	if discord.Client != nil && discord.Client.DataReady && nickname != "" {
		// Get the bot channel
		botChannel, botChannelErr := discord.Client.Channel(os.Getenv("DISCORD_CHAT_CHANNEL_ID"))
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
		// discord.logger.Trace("Successfully updated the nickname:", string(updateNicknameResponse))
	} else {
		if discord.Client == nil {
			return errors.New("Can't update nickname, Discord client is nil (discord.Client = nil)")
		} else if !discord.Client.DataReady {
			return errors.New("Can't update nickname, Discord client is not ready (discord.Client.DataReady = false)")
		}
		return errors.New("Can't update nickname, unknown issue with Discord client")
	}

	return nil
}

func (discord *Discord) updatePresence(presence string) error {
	if !discord.IsReady {
		return errors.New("Can't update presence, Discord not ready (discord.IsReady = false)")
	}

	// Set the presence
	if discord.Client != nil && discord.Client.DataReady && presence != "" {
		if statusErr := discord.Client.UpdateGameStatus(0, presence); statusErr != nil {
			discord.HasPresence = false
			return statusErr
		}
		discord.HasPresence = true
	} else {
		if discord.Client == nil {
			return errors.New("Can't update presence, Discord client is nil (discord.Client = nil)")
		} else if !discord.Client.DataReady {
			return errors.New("Can't update presence, Discord client is not ready (discord.Client.DataReady = false)")
		}
		return errors.New("Can't update presence, unknown issue with Discord client")
	}

	return nil
}
