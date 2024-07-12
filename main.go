package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// BasicAuthMiddleware is a middleware function for HTTP Basic Authentication
func BasicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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

func main() {
	username := "admin"
	password := "password"

	hub1 := NewHub("h1", "model3", "hub01")
	d1 := NewDevice("d1", "modelX", "device01")
	d2 := NewDevice("d2", "modelY", "device02")
	d3 := NewDevice("d3", "modelS", "device03")

	hub1.AddChild(d1)
	hub1.AddChild(d2)
	hub1.AddChild(d3)

	println("hub1.Tag()", hub1.Tag())
	println("d1.Tag()", d1.Tag())
	println("d2.Tag()", d2.Tag())
	println("d3.Tag()", d3.Tag())

	http.Handle("/ws", BasicAuthMiddleware(websocket.Handler(WebSocketHandler), username, password))
	http.Handle("/", BasicAuthMiddleware(hub1, username, password))

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
