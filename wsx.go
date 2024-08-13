package thing2

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// wsxHandle handles /wsx requests on an htmx WebSocket
func wsxHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Install /wsx websocket listener")
	serv := websocket.Server{Handler: websocket.Handler(wsxServe)}
	serv.ServeHTTP(w, r)
}

// wsxServe handles htmx WebSocket connections
func wsxServe(ws *websocket.Conn) {

	defer ws.Close()

	req := ws.Request()
	sessionId := req.URL.Query().Get("session-id")
	if sessionId == "" {
		println("missing session-id param")
		return
	}

	sessionConn(sessionId, ws)
	sessionDeviceSaveView(sessionId, root.Id, "full")
	sessionDeviceRender(sessionId, root.Id)

	var message string
	for {
		// Read message from the client
		if err := websocket.Message.Receive(ws, &message); err != nil {
			fmt.Println("Can't receive:", err)
			break
		}
		fmt.Println("Received message from client:", message)
	}

	sessionConn(sessionId, nil)
}
