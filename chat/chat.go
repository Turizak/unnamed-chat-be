package chat

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	ChannelID string `json:"channelId"`
	Username  string `json:"username"`
	Message   string `json:"message"`
}

type Channel struct {
	clients map[*websocket.Conn]bool
}

var (
	mu       sync.Mutex
	channels = make(map[string]*Channel)
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Get the channel ID from the query params
	channelID := r.URL.Query().Get("channelId")
	if channelID == "" {
		http.Error(w, "Missing channelId", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// Register the connection to the channel
	mu.Lock()
	if _, exists := channels[channelID]; !exists {
		channels[channelID] = &Channel{
			clients: make(map[*websocket.Conn]bool),
		}
	}
	channels[channelID].clients[conn] = true
	mu.Unlock()

	fmt.Printf("New connection in channel %s\n", channelID)

	// Handle incoming messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			break // Disconnect on error
		}

		msg.ChannelID = channelID
		BroadcastToChannel(channelID, msg)
	}

	// Remove the connection when the client disconnects
	mu.Lock()
	delete(channels[channelID].clients, conn)
	if len(channels[channelID].clients) == 0 {
		delete(channels, channelID) // Remove empty channels
	}
	mu.Unlock()
}

func BroadcastToChannel(channelID string, msg Message) {
	mu.Lock()
	defer mu.Unlock()

	clients, exists := channels[channelID]
	if !exists {
		return
	}

	for client := range clients.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(clients.clients, client)
		}
	}
}
