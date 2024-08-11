package thing2

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var (
	port         = GetEnv("PORT", "8000")
	deployParams = GetEnv("DEPLOY_PARAMS", "")
	root         *Device
)

func Run() {

	err := fileReadDevices()
	if err != nil {
		fmt.Println("Error reading devices from file:", err)
		return
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

	// (Optional) run as web server, if port given
	if port == "" {
		return
	}

	// Install /model/{model} patterns for makers
	modelsInstall()

	// Install the /device/{id} pattern for devices
	devicesInstall()

	// Install / to point to root device
	http.Handle("/", basicAuthHandler(root))

	// Install /ws websocket listener
	http.HandleFunc("/ws", basicAuthHandlerFunc(ws))

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
