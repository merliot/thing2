package thing2

import (
	"html/template"
	"log"
	"net/http"
)

var rootDevicer Devicer

var port = GetEnv("PORT", "8000")

func Run(root Devicer) {

	rootDevicer = root

	// Build route table from root's perpective
	BuildRoutes(root)

	// Dial parents
	dialParents()

	// (Optional) run as web server
	if port == "" {
		return
	}

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

	// Install /server/* patterns for debug info
	http.HandleFunc("/server/sessions", basicAuthHandlerFunc(sessionsShow))
	http.HandleFunc("/server/routes", basicAuthHandlerFunc(routesShow))

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
