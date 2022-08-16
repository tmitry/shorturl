package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func NewServer() *http.Server {
	router := mux.NewRouter()

	rep := repositories.NewMemoryRepository()

	shortenerHandler := handlers.ShortenerHandler{
		Rep: rep,
	}
	router.Handle("/", shortenerHandler).Methods("POST")

	redirectHandler := handlers.RedirectHandler{
		Rep: rep,
	}
	router.Handle(fmt.Sprintf("/{id:%s}", config.PatternUID), redirectHandler).Methods("GET")

	server := &http.Server{
		Addr:              config.DefaultAddr,
		Handler:           router,
		ReadHeaderTimeout: config.ServerReadHeaderTimeout * time.Second,
	}

	return server
}

func StartServer() {
	server := NewServer()

	log.Fatal(server.ListenAndServe())
}
