package main

import (
	"database/sql"
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

	messages []string

	db *sql.DB
}

func retrieveMessages(db *sql.DB) []string {
	rows, err := db.Query("SELECT message FROM messages")
	if err != nil {
		log.Println("Error retrieving messages from the database:", err)
		return []string{}
	}
	defer rows.Close()

	var messages []string

	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			log.Println("Error scanning message from database:", err)
			continue
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over messages:", err)
	}

	return messages
}

func (r *room) saveMessageToDB(msg string) {
	_, err := r.db.Exec("INSERT INTO messages (message) VALUES (?)", msg)
	if err != nil {
		log.Println("Error saving message to database:", err)
	}
}

// newRoom create a new chat room
func newRoom(db *sql.DB) *room {
	// get messages, if any, from database
	messages := retrieveMessages(db)

	return &room{
		forward:  make(chan []byte),
		join:     make(chan *client),
		leave:    make(chan *client),
		clients:  make(map[*client]bool),
		messages: messages,
		db:       db,
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			for _, msg := range r.messages {
				client.receive <- []byte(msg)
			}
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.receive)
		case msg := <-r.forward:
			r.saveMessageToDB(string(msg))
			r.messages = append(r.messages, string(msg))
			for client := range r.clients {
				client.receive <- msg
			}
		}
	}
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
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
