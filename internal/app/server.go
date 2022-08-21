package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func NewRouter() http.Handler {
	router := chi.NewRouter()

	rep := repositories.NewMemoryRepository()

	shortenerHandler := handlers.NewShortenerHandler(rep)
	router.Post("/", shortenerHandler.ServeHTTP)

	redirectHandler := handlers.NewRedirectHandler(rep)
	router.Get(fmt.Sprintf("/{%s:%s}", handlers.ParameterNameUID, config.PatternUID), redirectHandler.ServeHTTP)

	return router
}

func NewServer(router http.Handler) *http.Server {
	server := &http.Server{
		Addr:              config.DefaultAddr,
		Handler:           router,
		ReadHeaderTimeout: config.ServerReadHeaderTimeout * time.Second,
	}

	return server
}

func StartServer() {
	router := NewRouter()
	server := NewServer(router)

	log.Fatal(server.ListenAndServe())
}
