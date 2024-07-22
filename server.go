package thing2

import (
	"log"
	"net/http"
)

func Run(root Devicer, addr string) {

	// Install / to point to root device
	http.Handle("/", BasicAuth(root))

	// Install the /device/{id} pattern for root device
	root.InstallDevicePattern()

	// Install the /model/{model} pattern, using root device as proto (but
	// only if we haven't seen this model before)
	root.InstallModelPattern()

	println("ListenAndServe on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
