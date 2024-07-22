package thing2

import (
	"fmt"
	"io"
	"net/http"
)

type pageVars map[string]any

type pageData struct {
	Vars   pageVars
	Device any
}

func (d *Device) render(w io.Writer, name string, data any) error {
	tmpl := d.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("Template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}

func (d *Device) renderTemplateData(w http.ResponseWriter, name string, data any) {
	if err := d.render(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (d *Device) renderPage(w http.ResponseWriter, name string, pageVars pageVars) {
	d.renderTemplateData(w, name, &pageData{
		Vars:   pageVars,
		Device: d.data,
	})
}

func (d Device) showIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.renderPage(w, "index.tmpl", pageVars{
			"view": "full",
		})
	})
}
