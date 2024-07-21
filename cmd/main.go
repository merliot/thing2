package main

import (
	"fmt"

	"github.com/merliot/thing2"
	"github.com/merliot/thing2/gadget"
	"github.com/merliot/thing2/hub"
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

func main() {
	thing2.User = "user"
	thing2.Passwd = "passwd"
	addr := ":8080"

	hub1 := hub.NewHub("h1", "model3", "hub01")

	g1 := gadget.NewGadget("g1", "modelX", "gadget01")
	g2 := gadget.NewGadget("g2", "modelY", "gadget02")
	g3 := gadget.NewGadget("g3", "modelS", "gadget03")

	g4 := gadget.NewGadget("g4", "modelS", "gadget04")
	g5 := gadget.NewGadget("g5", "modelS", "gadget05")

	hub1.AddChild(g1.Device)
	hub1.AddChild(g2.Device)
	hub1.AddChild(g3.Device)

	g4.AddChild(g5.Device)
	hub1.AddChild(g4.Device)

	server := thing2.NewServer(hub1.Device, addr)
	server.Run()
}
