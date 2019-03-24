package discord

import (
	"errors"

	"github.com/Dids/rustbot/webrcon"
)

func (discord *Discord) updatePlayers(players []webrcon.PlayerPacket) error {
	if !discord.IsReady {
		return errors.New("Can't update nickname, Discord not ready (discord.IsReady = false)")
	}

	// TODO: Actually implement updating the player list "static message"
	discord.logger.Trace("Received players:", players)

	return nil
}
