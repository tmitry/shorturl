package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const messageIncorrectURL = "incorrect URL"

type ShortenerHandler struct {
	Rep repositories.Repository
}

func (h ShortenerHandler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	url := models.URL(body)

	if !url.IsValid() {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			http.StatusBadRequest,
		)

		return
	}

	shortURL := h.Rep.Add(url)

	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(shortURL.GetShortURL()))

	if err != nil {
		log.Println(err.Error())
	}
}
