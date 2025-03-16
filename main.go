package main

import (
	"fmt"
	"net/http"
	"unnamed-chat/chat"
)

func main() {
	http.HandleFunc("/ws", chat.HandleConnections)

	port := "8080"
	fmt.Printf("Starting WebSocket server on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
