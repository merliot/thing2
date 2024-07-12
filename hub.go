package main

type Hub struct {
	*Device
}

func NewHub(id, model, name string) *Hub {
	return &Hub{
		Device: NewDevice(id, model, name),
	}
}
