package webrcon

import "github.com/sacOO7/gowebsocket"

func (webrcon *Webrcon) handlePongReceived(message string, socket gowebsocket.Socket) {
	if webrcon.isShuttingDown {
		webrcon.logger.Warning("Already shutting down!")
		return
	}
}
