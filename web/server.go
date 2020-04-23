package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Settings struct {
	Port           string
	HeaderTimeout  time.Duration
	RequestTimeout time.Duration
	WriteTimeout   time.Duration
}

func DefaultSettings() *Settings {
	return &Settings{
		Port:           "8000",
		HeaderTimeout:  time.Second * 10,
		RequestTimeout: time.Second * 10,
		WriteTimeout:   time.Second * 10,
	}
}

type Server struct {
	router   *mux.Router
	settings *Settings
}

func NewServer(s *Settings, api *Api) *Server {
	router := mux.NewRouter()
	router.StrictSlash(true)
	registerControllers(api, router)

	return &Server{
		router:   router,
		settings: s,
	}
}

func registerControllers(api *Api, router *mux.Router) {
	for _, ctrl := range api.Controllers {
		for _, route := range ctrl.Routes() {
			log.Printf("Registering endpoint: %s %s", route.Endpoint.Method, route.Endpoint.Path)
			router.Handle(route.Endpoint.Path, route.Handler).Methods(route.Endpoint.Method)
		}
	}
}

func (s *Server) Run(ctx context.Context) {
	server := &http.Server{
		Handler:           s.router,
		Addr:              "127.0.0.1:" + s.settings.Port,
		ReadTimeout:       s.settings.RequestTimeout,
		WriteTimeout:      s.settings.WriteTimeout,
		ReadHeaderTimeout: s.settings.HeaderTimeout,
	}
	log.Printf("Server listening on port %s...", s.settings.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Panicf("http server failed: %s", err)
	}
}
