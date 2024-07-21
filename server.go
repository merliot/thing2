package thing2

import (
	"log"
	"net/http"
)

type Server struct {
	http.Server
	user   string
	passwd string
}

func NewServer(root *Device, addr string) *Server {
	s := &Server{
		Server: http.Server{
			Addr: addr,
		},
	}

	http.Handle("/", BasicAuth(root))
	root.HandleDevice()

	return s
}

func (s Server) Run() {
	println("ListenAndServe on", s.Server.Addr)
	log.Fatal(s.ListenAndServe())
}
