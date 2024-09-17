//go:build !tinygo

package thing2

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

var root *Device

func Run() {

	var port = Getenv("PORT", "8000")
	var demo = (Getenv("DEMO", "") == "true")
	var err error

	if err := devicesLoad(); err != nil {
		fmt.Println("Error loading devices:", err)
		return
	}

	devicesMake()

	root, err = devicesFindRoot()
	if err != nil {
		fmt.Println("Error finding root device:", err)
		return
	}

	if demo {
		root.Set(flagDemo)
	}

	if err := root.Setup(); err != nil {
		fmt.Println("Error setting up root device:", err)
		return
	}

	// Build route table from root's perpective
	routesBuild(root)

	// Dial parents
	dialParents()

	// If no port was given, don't run as a web server
	if port == "" {
		root.run()
		log.Println("Device", root.Name, "done, bye")
		return
	}

	// Running as a web server...

	// Install /model/{model} patterns for makers
	modelsInstall()

	// Install the /device/{id} pattern for devices
	devicesInstall()

	// Install / to point to root device
	http.Handle("/", basicAuthHandler(root))

	// Install /ws websocket listener
	http.HandleFunc("/ws", basicAuthHandlerFunc(wsHandle))

	// Install /wsx websocket listener (wsx is for htmx ws)
	http.HandleFunc("/wsx", basicAuthHandlerFunc(wsxHandle))

	// Install /server/* patterns for debug info
	http.HandleFunc("/server/sessions", basicAuthHandlerFunc(sessionsShow))

	addr := ":" + port
	server := &http.Server{Addr: addr}

	// Run http server in go routine to be shutdown later
	go func() {
		fmt.Println("ListenAndServe on", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}

	}()

	root.run()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	log.Println("Device", root.Name, "done, bye")
}
