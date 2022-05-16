package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
)

func (discord *Discord) handleConnect(session *discordgo.Session, event *discordgo.Connect) {
	discord.logger.Trace("Discord event: connect")
	discord.IsReady = true
}

func (discord *Discord) handleDisconnect(session *discordgo.Session, event *discordgo.Disconnect) {
	discord.logger.Trace("Discord event: disconnect")
	discord.IsReady = false

	// Notify the primary process to shut down
	discord.logger.Trace("Shutting down!")
	process, _ := os.FindProcess(os.Getpid())
	process.Signal(os.Interrupt)
	return
}

func (discord *Discord) handleRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	discord.logger.Trace("Discord event: ratelimit")
}

func (discord *Discord) handleReady(session *discordgo.Session, event *discordgo.Ready) {
	discord.logger.Trace("Discord event: ready")
	discord.IsReady = true
}
