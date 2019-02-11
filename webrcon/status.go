package webrcon

import "time"

// TODO: Rewrite this at some point so we can stop it if we want to,
//       for example if we start updating after the "ready" websocket
//       event, as that's where it would make the most sense, right?
func (webrcon *Webrcon) startUpdatingStatus() {
	defer func() {
		webrcon.logger.Trace("Stopping status updater..")
	}()

	for {
		if webrcon.isShuttingDown {
			return
		}

		webrcon.logger.Trace("Requesting status..")

		// Lock the write mutex
		webrcon.writeMutex.Lock()

		// Set the deadline (timeout) for reading the next incoming message
		//webrcon.Client.Conn.SetReadDeadline(time.Now().Add(pongWait)) // TODO: Since this is pong specific, we shouldn't use it here, but create a const time for statusWait instead?

		// Set the deadline (timeout) for the next sent message
		webrcon.Client.Conn.SetWriteDeadline(time.Now().Add(writeWait))

		// Send the status request message
		webrcon.Client.SendText(`{ "Message": "status", "Identifier": 0, "Type": "Generic" }`)

		// Unlock the write mutex
		webrcon.writeMutex.Unlock()

		// Sleep for a bit before requesting the status (Discord API only allows the presence to be updated every 15 seconds)
		time.Sleep(15 * time.Second)
	}
}
