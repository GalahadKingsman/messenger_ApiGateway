package server

import (
	"log"
	"messenger_frontend/internal/config"
	"net/http"
)

type Server struct {
	cfg      *config.Config
	handlers *handlers.Handler
}

func New(cfg *config.Config) *Server {
	return &Server{
		cfg:      cfg,
		handlers: handlers.New(cfg),
	}
}

func (s *Server) Run() error {
	http.HandleFunc("/users", s.handlers.GetUser)
	http.HandleFunc("/dialogs", s.handlers.GetDialog)

	log.Printf("Starting HTTP server on %s", s.cfg.HTTP.Port)
	return http.ListenAndServe(":"+s.cfg.HTTP.Port, nil)
}
