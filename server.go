package thing2

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var root *Device

func Run() {

	var port = getenv("PORT", "8000")
	var demo = (getenv("DEMO", "") == "true")
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
	fmt.Println("ListenAndServe on", addr)
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
