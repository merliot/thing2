package thing2

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// ws handles /ws requests on a WebSocket
func ws(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsServe)}
	serv.ServeHTTP(w, r)
}

// wsServe handles WebSocket connections
func wsServe(ws *websocket.Conn) {
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
