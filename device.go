package main

import (
	"fmt"
	"net/http"
)

type Device struct {
	*Thing
	*http.ServeMux
	Model string
}

func NewDevice(id, model, name string) *Device {
	d := &Device{
		Thing:    NewThing(id, name),
		ServeMux: http.NewServeMux(),
		Model:    model,
	}
	return d
}

func (d *Device) handle(thinger Thinger) {
	println("/device/" + thinger.GetId())
	//d.Handle("/device/"+thinger.Tag(), thinger)
	println("/device/" + thinger.GetId() + "/")
	//d.Handle("/device/"+thinger.Tag()+"/", thinger)
}

func (d *Device) AddChild(child Thinger) error {
	if err := d.Thing.AddChild(child); err != nil {
		return err
	}
	d.handle(child)
	return nil
}

func (d *Device) xServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hello from Id: %s Tag: %s", d.Id, d.Tag())
}
