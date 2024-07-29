package thing2

import (
	"log"
	"net/http"
)

func Run(root Devicer, addr string) {

	// Install / to point to root device
	http.Handle("/", basicAuthHandler(root))

	// Install the /device/{id} pattern for root device
	root.InstallDevicePattern()

	// Install the /model/{model} pattern, using root device as proto (but
	// only if we haven't seen this model before)
	root.InstallModelPattern()

	// Install /ws websocket listener
	http.HandleFunc("/ws", basicAuthHandlerFunc(ws))

	// Install /wsx websocket listener
	http.HandleFunc("/wsx", basicAuthHandlerFunc(wsxHandle))

	// Build route table from root's perpective
	BuildRoutes(root)

	println("ListenAndServe on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
