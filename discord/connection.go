package discord

import (
	"github.com/bwmarrin/discordgo"
)

// TODO: Is this called on reconnect? Because if not, we should NOT set "discord.IsReady = false"
//       when we get disconnected, as automatic reconnection etc. should handle everything for us!
func (discord *Discord) handleConnect(session *discordgo.Session, event *discordgo.Connect) {
	discord.logger.Trace("Discord event: connect")
	discord.IsReady = true
}

func (discord *Discord) handleDisconnect(session *discordgo.Session, event *discordgo.Disconnect) {
	discord.logger.Trace("Discord event: disconnect")

	// TODO: If discordgo keeps buffering our messages and stuff, we would ideally NOT want to set this,
	//       because we want our data to keep flowing, even if it's buffered and going in/out "late"!

	// discord.IsReady = false // TODO: Test if we need to set this or not, or if discordgo keeps buffering data for us?

	// TODO: We're no longer shutting down when Discord disconnects,
	//       as the client will keep reconnecting, so we will just keep
	//       running and waiting for the client to reconnect.

	// Notify the primary process to shut down
	// discord.logger.Trace("Shutting down!")
	// process, _ := os.FindProcess(os.Getpid())
	// process.Signal(os.Interrupt)
	// return
}

func (discord *Discord) handleRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	discord.logger.Trace("Discord event: ratelimit")
}

func (discord *Discord) handleReady(session *discordgo.Session, event *discordgo.Ready) {
	discord.logger.Trace("Discord event: ready")
	discord.IsReady = true
}
