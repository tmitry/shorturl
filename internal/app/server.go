package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const (
	ContextKeyUserID middlewares.ContextKey = "userID"

	jwtCookieName = "jwt"
)

func NewRouter(cfg *configs.Config) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Compress(cfg.Server.CompressionLevel))
	router.Use(middlewares.JWTAuth(cfg.Server.JWTSignatureKey, jwtCookieName, ContextKeyUserID))

	var rep repositories.Repository

	if cfg.App.FileStoragePath != "" {
		rep = repositories.NewFileRepository(cfg.App.FileStoragePath)
	} else {
		rep = repositories.NewMemoryRepository()
	}

	shortenerHandler := handlers.NewShortenerHandler(cfg, rep, ContextKeyUserID)
	shortenerAPIHandler := handlers.NewShortenerAPIHandler(cfg, rep, ContextKeyUserID)

	router.Route("/", func(router chi.Router) {
		router.Use(middleware.AllowContentType(handlers.ContentTypeText, handlers.ContentTypeGZIP))
		router.Use(middleware.AllowContentEncoding(handlers.ContentEncodingGZIP))
		router.Post("/", shortenerHandler.Shorten)
		router.Get(fmt.Sprintf(
			"/{%s:%s}",
			handlers.ParameterNameUID,
			models.GetPatternUID(cfg.App.HashMinLength),
		), shortenerHandler.Redirect)
		router.Get("/ping", shortenerHandler.Ping)
	})

	router.Route("/api", func(router chi.Router) {
		router.Use(middleware.AllowContentType(handlers.ContentTypeJSON, handlers.ContentTypeGZIP))
		router.Use(middleware.AllowContentEncoding(handlers.ContentEncodingGZIP))
		router.Post("/shorten", shortenerAPIHandler.Shorten)
		router.Get("/user/urls", shortenerAPIHandler.UserUrls)
	})

	return router
}

func NewServer(router http.Handler, serverCfg *configs.ServerConfig) *http.Server {
	server := &http.Server{
		Addr:              serverCfg.Address,
		Handler:           router,
		ReadHeaderTimeout: time.Duration(serverCfg.ReadHeaderTimeout) * time.Second,
	}

	return server
}

func StartServer(cfg *configs.Config) {
	router := NewRouter(cfg)
	server := NewServer(router, cfg.Server)

	log.Fatal(server.ListenAndServe())
}
