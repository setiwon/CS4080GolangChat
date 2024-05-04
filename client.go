package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user.

type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn

	// receive is a channel to receive messages from other clients.
	receive chan []byte

	// room is the room this client is chatting in.
	room *room

	// name of the client
	name string
}

type clientMessage struct {
	Method string
	Body   string
}

// received message from this client
func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, recvErr := c.socket.ReadMessage()
		if recvErr != nil {
			return
		}

		fmt.Printf("%s\n", msg)

		var parsed clientMessage
		parseErr := json.Unmarshal(msg, &parsed)
		if parseErr != nil {
			fmt.Printf("Message parse error: %s\n", msg)
			// if parseErr we can just ignore the message
		} else {
			fmt.Printf("%s %s\n", parsed.Method, parsed.Body)
			switch parsed.Method {
			case "name":
				c.name = parsed.Body
			case "message":
				encodedMessage, _ := json.Marshal(map[string]string{"sender": c.name, "message": parsed.Body})
				c.room.forward <- encodedMessage
			}
		}

	}
}

// received message from other clients, send to this client
func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
