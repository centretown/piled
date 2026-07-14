package socket

import (
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	mutex           sync.Mutex
	hub             *Hub
	statusLayout    *template.Template
	activeRecorders int
}

func NewServer(t *template.Template) *Server {
	s := &Server{
		hub:          NewHub(),
		statusLayout: t.Lookup("layout.wsstatus"),
	}
	return s
}
func (s *Server) Run() {
	go s.hub.Run()
}

func (s *Server) Broadcast(message string) {
	s.hub.Broadcast(message)
}

func (s *Server) Events(w http.ResponseWriter, r *http.Request) {
	client, err := NewClient(s.hub, w, r)
	if err != nil {
		log.Printf("Failed to create WebSocket client: %v", err)
		return
	}

	s.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) Done() {
	s.hub.Done <- true
}
