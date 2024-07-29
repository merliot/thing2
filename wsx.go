package thing2

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// wsxHandle handles /wsx requests on an htmx WebSocket
func wsxHandle(w http.ResponseWriter, r *http.Request) {
	serv := websocket.Server{Handler: websocket.Handler(wsxServe)}
	serv.ServeHTTP(w, r)
}

// wsxServe handles htmx WebSocket connections
func wsxServe(ws *websocket.Conn) {

	defer ws.Close()

	req := ws.Request()
	id := req.URL.Query().Get("session-id")
	if id == "" {
		println("missing session-id param")
		return
	}

	sessionConn(id, ws)

	// TODO send /full rendering back over websocket

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

	sessionConn(id, nil)
}

func BcastUp(path string, msg any) {

}
