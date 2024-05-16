package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// clients keeps track of the room each connection belongs to
var clients = make(map[string][]*websocket.Conn)

//var clients = make(map[*websocket.Conn]string)

// broadcast is a channel for sending messages to specific rooms
var broadcast = make(chan Message)

// Message defines the structure for a user's message along with their room
type Message struct {
	Username string
	Message  string
	Room     string
}

// upgrader is used to upgrade the HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connection from any origin
	},
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Get username and room from the query parameters
	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")

	// Upgrade the HTTP server connection to the WebSocket protocol
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}
	defer conn.Close()

	if _, ok := clients[room]; !ok {
		clients[room] = []*websocket.Conn{conn}
	} else {
		clients[room] = append(clients[room], conn)
	}

	fmt.Println(clients)

	fmt.Printf("%s joined %s\n", username, room)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, room)
			break
		}

		broadcast <- Message{Username: username, Message: string(message), Room: room}
	}
}
func handleMessages() {
	for {
		msg := <-broadcast

		log.Printf("%s in %s: %s\n", msg.Username, msg.Room, msg.Message)

		for _, conn := range clients[msg.Room] {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Username+": "+msg.Message))
			if err != nil {
				log.Printf("Websocket error: %v", err)
				conn.Close()
			}
		}
	}
}
