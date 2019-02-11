package webrcon

import (
	"os"
	"time"

	"github.com/Dids/rustbot/eventhandler"
	"github.com/gorilla/websocket"
	"github.com/sacOO7/gowebsocket"
)

func (webrcon *Webrcon) startPinging() {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}

	webrcon.logger.Trace("Creating PING ticker..")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		webrcon.logger.Trace("Stopping PING ticker..")
		ticker.Stop()

		// Send server disconnected message to Discord
		webrcon.EventHandler.Emit(eventhandler.Message{Event: "receive_webrcon_message", User: "", Message: "Disconnected from server!", Type: eventhandler.ServerDisconnectedType})

		// Sleep for a bit before shutting down
		time.Sleep(1 * time.Second)

		// Notify the primary process to shut down
		webrcon.logger.Error("Disconnected from server, shutting down..")
		process, _ := os.FindProcess(os.Getpid())
		process.Signal(os.Interrupt)
		return
	}()
	for {
		if webrcon.isShuttingDown {
			webrcon.logger.Warning("Already shutting down!")
			return
		}

		select {
		case <-ticker.C:
			// Send PING
			webrcon.writeMutex.Lock()
			webrcon.Client.Conn.SetReadDeadline(time.Now().Add(pongWait)) // TODO: This might break things, or simply be unnecessary
			webrcon.Client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := webrcon.Client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				webrcon.logger.Error("Failed to send PING to server:", err)

				// Remember to unblock, just in case
				webrcon.writeMutex.Unlock()

				// Return so that the ticker is stopped
			}
			webrcon.writeMutex.Unlock()
		}
	}
}

func (webrcon *Webrcon) handlePingReceived(message string, socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}
}
