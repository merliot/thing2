package server

import (
	"log"
	"net/http"

	"github.com/merliot/thing2/device"
)

type Server struct {
	http.Server
	user string
	passwd string
}

func NewServer(root *device.Device, addr string) *Server {
	s := &Server{
		Server: http.Server{
			Addr: addr,
		},
	}

	http.Handle("/", device.BasicAuth(root))
	root.HandlePrefix()

	return s
}

func (s Server) Run() {
	println("ListenAndServe on", s.Server.Addr)
	log.Fatal(s.ListenAndServe())
}
