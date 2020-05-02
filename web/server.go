package web

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type Settings struct {
	Host              string        `mapstructure:"host"`
	Port              string        `mapstructure:"port"`
	HeaderTimeout     time.Duration `mapstructure:"header_timeout"`
	RequestTimeout    time.Duration `mapstructure:"request_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	UseCSRFProtection bool          `mapstructure:"use_csrf_protection"`
	CSRFTokenKey      string        `mapstructure:"csrf_key"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Host:              "",
		Port:              "8000",
		HeaderTimeout:     time.Second * 10,
		RequestTimeout:    time.Second * 10,
		WriteTimeout:      time.Second * 10,
		CSRFTokenKey:      "ffffff",
		UseCSRFProtection: true,
	}
}

func (s *Settings) Keys() []string {
	return []string{
		"host",
		"port",
		"header_timeout",
		"request_timeout",
		"write_timeout",
		"use_csrf_protection",
		"csrf_key",
	}
}

type Server struct {
	Router   *mux.Router
	settings *Settings
}

func NewServer(s *Settings, api *Api) *Server {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.Use(recoveryMiddleware())
	if s.UseCSRFProtection {
		router.Use(csrf.Protect([]byte(s.CSRFTokenKey), csrf.Secure(false)))
	}
	registerControllers(api, router)

	return &Server{
		Router:   router,
		settings: s,
	}
}

func registerControllers(api *Api, router *mux.Router) {
	for _, ctrl := range api.Controllers {
		for _, route := range ctrl.Routes() {
			log.Printf("Registering endpoint: %s %s", route.Endpoint.Method, route.Endpoint.Path)
			securedHandler := ScopeWrapper(route.Handler, route.Scopes)
			wrappedHandler := HandlerWrapper(securedHandler, route.ModelBlueprint)
			chainedHandler := Chain(wrappedHandler, MatchFilters(route.Endpoint, api.Filters))
			router.Handle(route.Endpoint.Path, chainedHandler).Methods(route.Endpoint.Method)
		}
	}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	server := &http.Server{
		Handler:           s.Router,
		Addr:              s.settings.Host + ":" + s.settings.Port,
		ReadTimeout:       s.settings.RequestTimeout,
		WriteTimeout:      s.settings.WriteTimeout,
		ReadHeaderTimeout: s.settings.HeaderTimeout,
	}
	go shutdownServer(ctx, wg, server)

	wg.Add(1)
	log.Printf("Server listening on %s:%s...", s.settings.Host, s.settings.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Panicf("http server failed: %s", err)
	}
}

func shutdownServer(ctx context.Context, wg *sync.WaitGroup, server *http.Server) {
	<-ctx.Done()
	defer wg.Done()
	log.Printf("Server shutting down. Will wait for 5 seconds...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Server errored while shutting down: %s", err)
	}
}
