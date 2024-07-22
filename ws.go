package thing2

import (
	"fmt"

	"golang.org/x/net/websocket"
)

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(ws *websocket.Conn) {
	defer ws.Close()
	var message string
	for {
		// Read message from the client
		if err := websocket.Message.Receive(ws, &message); err != nil {
			fmt.Println("Can't receive:", err)
			break
		}
		fmt.Println("Received message from client:", message)

		// Send message back to the client
		if err := websocket.Message.Send(ws, "Echo: "+message); err != nil {
			fmt.Println("Can't send:", err)
			break
		}
	}
}
