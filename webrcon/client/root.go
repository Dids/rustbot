package client

import (
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

// rust.suomileijona.eu
// 28016
// LywCx4kAgEFdtYTGkaGw9wBXMvijTsBJ

var client *websocket.Conn

// Initialize creates and connects a new websocket connection
func Initialize() {
	// Specify the connection URL
	connectionURL := url.URL{Scheme: "ws", Host: os.Getenv("WEBRCON_HOST") + ":" + os.Getenv("WEBRCON_PORT"), Path: "/" + os.Getenv("WEBRCON_PASSWORD")}
	log.Printf("Websocket connecting to %s", connectionURL.String())

	// Create and start the websocket client/connection
	c, _, connectionErr := websocket.DefaultDialer.Dial(connectionURL.String(), nil)
	if connectionErr != nil {
		log.Fatal("Websocket connection failed:", connectionErr)
	}
	client = c
	log.Println("Websocket connection successful!")

	// FIXME: From what I can tell, this does absolutely nothing, unless we're meant to run this _with_ a timer or something?
	// Setup incoming message handling
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := client.ReadMessage()
			if err != nil {
				log.Println("Read:", err)
				return
			}
			log.Printf("Received: %s", message)
		}
	}()
}

// Close cleans up the websocket connections
func Close() {
	log.Println("Closing websocket connection..")
	// TODO: Ideally we'd wait for the server to close the connection, falling back to a timeout
	err := client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Failed to send websocket close message:", err)
	}
	client.Close()
	log.Println("Successfully closed the websocket connection!")
}
