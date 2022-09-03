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
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func NewRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Compress(config.ServerCfg.CompressionLevel))

	var rep repositories.Repository

	if config.AppCfg.FileStoragePath != "" {
		rep = repositories.NewFileRepository()
	} else {
		rep = repositories.NewMemoryRepository()
	}

	shortenerHandler := handlers.NewShortenerHandler(rep)
	shortenerAPIHandler := handlers.NewShortenerAPIHandler(rep)

	router.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handlers.ContentTypeText, handlers.ContentTypeGZIP))
		r.Use(middleware.AllowContentEncoding(handlers.ContentEncodingGZIP))
		r.Post("/", shortenerHandler.Shorten)
		r.Get(fmt.Sprintf("/{%s:%s}", handlers.ParameterNameUID, models.GetPatternUID()), shortenerHandler.Redirect)
	})

	router.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handlers.ContentTypeJSON, handlers.ContentTypeGZIP))
		r.Use(middleware.AllowContentEncoding(handlers.ContentEncodingGZIP))
		r.Post("/shorten", shortenerAPIHandler.Shorten)
	})

	return router
}

func NewServer(router http.Handler) *http.Server {
	server := &http.Server{
		Addr:              config.ServerCfg.Address,
		Handler:           router,
		ReadHeaderTimeout: time.Duration(config.ServerCfg.ReadHeaderTimeout) * time.Second,
	}

	return server
}

func StartServer() {
	router := NewRouter()
	server := NewServer(router)

	log.Fatal(server.ListenAndServe())
}
