package github

import (
	"net/http"
	"time"

	"github.com/google/go-github/v18/github"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// EventHandler interface to allow hooking into GitHub events
type EventHandler interface {
	HandleEvent(event interface{})
}

// Server used to listen for and pass off GitHub events
type Server struct {
	webhookSecretKey string
	eventHandler     EventHandler
}

// NewServer creates a new instance of Server, taking in ServerOptions
func NewServer(options ...ServerOption) *Server {
	zerolog.TimeFieldFormat = ""

	s := Server{}
	for _, option := range options {
		option(&s)
	}
	return &s
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	var err error
	var event interface{}

	payload, err := github.ValidatePayload(r, []byte(s.webhookSecretKey))
	if err == nil {
		event, err = github.ParseWebHook(github.WebHookType(r), payload)
	}

	if err != nil {
		log.Error().Err(err).Msg("Error handling incoming GitHub event")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.eventHandler.HandleEvent(event)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Start starts the server
func (s *Server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.handleEvents).Methods("POST")
	r.HandleFunc("/", s.healthCheck).Methods("GET")

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      r,
		IdleTimeout:  time.Second * 60,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
	}

	log.Fatal().Err(srv.ListenAndServe())
}
