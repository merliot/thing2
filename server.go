package thing2

import (
	"html/template"
	"log"
	"net/http"
)

var (
	port         = GetEnv("PORT", "8000")
	deployParams = GetEnv("DEPLOY_PARAMS", "")

	rootDevicer Devicer
)

func Run(root Devicer) {

	root.SetDeployParams(deployParams)
	rootDevicer = root

	// Build route table from root's perpective
	BuildRoutes(root)

	// Dial parents
	dialParents()

	// (Optional) run as web server, if port given
	if port == "" {
		return
	}

	// Install /model/{model} patterns for makers
	modelsInstall()

	// Install / to point to root device
	http.Handle("/", basicAuthHandler(root))

	// Install the /device/{id} pattern for root device
	root.InstallDevice()

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
