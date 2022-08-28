package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func NewRouter() http.Handler {
	router := chi.NewRouter()

	rep := repositories.NewMemoryRepository()

	shortenerHandler := handlers.NewShortenerHandler(rep)
	shortenerAPIHandler := handlers.NewShortenerAPIHandler(rep)

	router.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handlers.ContentTypeText))
		r.Post("/", shortenerHandler.Shorten)
		r.Get(fmt.Sprintf("/{%s:%s}", handlers.ParameterNameUID, config.PatternUID), shortenerHandler.Redirect)
	})

	router.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handlers.ContentTypeJSON))
		r.Post("/shorten", shortenerAPIHandler.Shorten)
	})

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
