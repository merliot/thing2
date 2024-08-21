package thing2

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync/atomic"
)

var (
	root  *Device
	dirty atomic.Bool
)

func Run() {

	var port = getenv("PORT", "8000")
	var devicesFile = getenv("DEVICES_FILE", "devices.json")
	var devicesJSON = getenv("DEVICES", "")
	var err error

	if devicesJSON != "" {
		if err := json.Unmarshal([]byte(devicesJSON), &devices); err != nil {
			fmt.Printf("Error parsing devices: %s\n", err)
			return
		}
	} else {
		if err := fileReadJSON(devicesFile, &devices); err != nil {
			fmt.Println("Error reading devices from file:", err)
			return
		}
	}

	devicesMake()

	root, err = devicesFindRoot()
	if err != nil {
		fmt.Println("Error finding root device:", err)
		return
	}

	// Build route table from root's perpective
	routesBuild()

	// Dial parents
	dialParents()

	// If no port was given, don't run as a web server
	if port == "" {
		select {}
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

	// Install /wsx websocket listener
	http.HandleFunc("/wsx", basicAuthHandlerFunc(wsxHandle))

	// Install /server/* patterns for debug info
	http.HandleFunc("/server/sessions", basicAuthHandlerFunc(sessionsShow))

	addr := ":" + port
	println("ListenAndServe on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func templateShow(w http.ResponseWriter, temp string, data any) {
	tmpl, err := template.New("main").Parse(temp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
