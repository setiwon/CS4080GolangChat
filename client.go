package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"html"
	"log"
	"os"
	"path/filepath"
	"strings"
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

type fileMetadata struct {
	Name string
	Type string
}

// received message from this client
func (c *client) read() {
	defer c.socket.Close()
	for {
		messageType, msg, recvErr := c.socket.ReadMessage()
		if recvErr != nil {
			return
		}

		if messageType == websocket.TextMessage {
			fmt.Printf("%s\n", msg)

			var parsed clientMessage
			parseErr := json.Unmarshal(msg, &parsed)
			if parseErr != nil {
				fmt.Printf("Message parse error: \n")
			} else {
				fmt.Printf("%s %s\n", parsed.Method, parsed.Body)
				switch parsed.Method {
				case "name":
					c.name = html.EscapeString(parsed.Body)
				case "message":
					escaped := html.EscapeString(parsed.Body)
					encodedMessage, _ := json.Marshal(map[string]string{"sender": c.name, "method": "message", "message": escaped})
					c.room.forward <- encodedMessage
				}
			}
		} else if messageType == websocket.BinaryMessage {
			metadataEndIdx := bytes.IndexByte(msg, byte('\x02'))
			if metadataEndIdx == -1 {
				log.Println("Metadata seperator not found.")
			}

			var metadata fileMetadata
			parseErr := json.Unmarshal(msg[:metadataEndIdx], &metadata)
			if parseErr != nil {
				log.Println("Metadata parse error.")
			} else {
				fmt.Printf("File received, name: '%s' type: '%s\n'", metadata.Name, metadata.Type)
				escapedName := html.EscapeString(metadata.Name)
				file := msg[metadataEndIdx+1:]
				// create filename based on hash
				hasher := sha256.New()
				hasher.Write(file)
				sha := base32.StdEncoding.EncodeToString(hasher.Sum(nil))
				// get file extension
				ext := escapedName[strings.LastIndex(escapedName, "."):]
				filePath := filepath.Join(FilesRoot, sha+ext)
				// write to "object store" aka disk
				writeErr := os.WriteFile(filePath, msg[metadataEndIdx+1:], 0644)
				if writeErr != nil {
					log.Println("File write error.")
				}

				fmt.Printf("File '%s' saved as '%s'", metadata.Name, filePath)

				encodedMessage, _ := json.Marshal(map[string]string{
					"sender": c.name, "method": "file", "name": escapedName, "path": filePath,
				})
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
