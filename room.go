package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {

	// clients holds all current clients in this room.
	clients map[*client]bool

	// join is a channel for clients wishing to join the room.
	join chan *client

	// leave is a channel for clients wishing to leave the room.
	leave chan *client

	// forward is a channel that holds incoming messages that should be forwarded to the other clients.
	forward chan []byte

	messages chan []byte // Channel to handle message persistence
}

// newRoom create a new chat room

func newRoom() *room {
	return &room{
		forward:  make(chan []byte),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[*client]bool),
		messages: make(chan []byte, 1000), // Buffered channel for storing recent messages
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.receive)
		case msg := <-r.forward:
			// Forward message to clients
			for client := range r.clients {
				client.receive <- msg
			}
			// Store message in the database
			r.messages <- msg
		}
	}
}

// load recent messagse from the database and send them to the client
func (r *room) loadMessages(client *client) {
	for msg := range r.messages {
		client.receive <- msg
	}
}

// close the message channel when the room is closed
func (r *room) close() {
	close(r.messages)
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
		name:    "",
	}
	r.join <- client
	r.loadMessages(client)
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
