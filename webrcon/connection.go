package webrcon

import (
	"os"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/sacOO7/gowebsocket"
)

func (webrcon *Webrcon) handleConnect(socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	webrcon.logger.Info("Connected to server")

	// Send server connected message to Discord
	webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "I'm back, baby!", Type: eventhandler.ServerConnectedType})
}

func (webrcon *Webrcon) handleDisconnect(err error, socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	webrcon.logger.Info("Disconnected from server", err)

	// When a disconnect error occurs, this means that we didn't gracefully shutdown, but the connection was lost etc.
	if err != nil {
		// Send server disconnected message to Discord
		webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Disconnected from server!", Type: eventhandler.ServerDisconnectedType})

		// Sleep for a bit before shutting down
		time.Sleep(1 * time.Second)

		// Notify the primary process to shut down
		webrcon.logger.Error("Disconnected from server, shutting down..")
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}
}

func (webrcon *Webrcon) handleConnectError(err error, socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	webrcon.logger.Info("Could not connect to server", err)

	if err != nil {
		// Send server disconnected message to Discord
		webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Cannot connect to server!", Type: eventhandler.ServerDisconnectedType})

		// Sleep for a bit before shutting down
		time.Sleep(1 * time.Second)

		// Notify the primary process to shut down
		webrcon.logger.Error("Could not connect to server, shutting down..")
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}
}
