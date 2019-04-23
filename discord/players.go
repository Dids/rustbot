package discord

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Dids/rustbot/webrcon"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/table"
)

func (discord *Discord) updatePlayers(players []webrcon.PlayerPacket) error {
	//discord.logger.Trace("Updating players:", players)

	if !discord.IsReady {
		return errors.New("Can't update player list, Discord not ready (discord.IsReady = false)")
	}

	// Skip if the the channel ID isn't set
	if len(os.Getenv("DISCORD_PLAYERLIST_CHANNEL_ID")) <= 0 {
		//discord.logger.Trace("DISCORD_PLAYERLIST_CHANNEL_ID not set, skipping player list update")
		return nil
	}

	// Get the player list channel
	playersChannel, err := discord.Client.Channel(os.Getenv("DISCORD_PLAYERLIST_CHANNEL_ID"))
	if err != nil {
		return err
	}

	// Generate the player list table string
	playersTable := table.NewWriter()
	playersTable.SetStyle(table.StyleLight)
	playersTable.AppendHeader(table.Row{"Username", "Ping", "Connected", "Violations", "Kicks"})
	for _, player := range players {
		//discord.logger.Trace("Parsing player:", player)

		// Skip invalid players
		if len(player.SteamID) > 0 {
			playerConnectedSeconds, err := strconv.ParseFloat(strings.Replace(player.Connected, "s", "", -1), 32)
			if err != nil {
				return err
			}
			playerConnectedTime := time.Now().Add(time.Duration(-playerConnectedSeconds) * time.Second)
			playersTable.AppendRow([]interface{}{player.Username, player.Ping, humanize.Time(playerConnectedTime), player.Violations, player.Kicks})
		}
	}
	playersMessage := "```\n"
	playersMessage += playersTable.Render()
	playersMessage += "\n```"

	// Check if any messages exist
	existingMessage, err := discord.Client.ChannelMessage(playersChannel.ID, playersChannel.LastMessageID)
	if err != nil {
		// Create a new message if one doesn't exist
		if _, err := discord.Client.ChannelMessageSend(playersChannel.ID, playersMessage); err != nil {
			return err
		}
	} else {
		// Skip if the message didn't change
		if existingMessage.Content == playersMessage {
			return nil
		}

		//discord.logger.Trace("existingMessage.Content:\n", existingMessage.Content)
		//discord.logger.Trace("playersMessage:\n", playersMessage)

		// Update the existing message if it already exists
		if _, err := discord.Client.ChannelMessageEdit(playersChannel.ID, existingMessage.ID, playersMessage); err != nil {
			return err
		}
	}

	return nil
}
