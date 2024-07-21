package device

import (
	"fmt"
	"net/http"
)

func (d *Device) showIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s showIndex path %s", d.Name, r.URL.Path)
	})
}
