package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
	"github.com/tmitry/shorturl/internal/app/util"
)

const messageIncorrectURL = "incorrect URL"

type ShortenerHandler struct {
	Rep repositories.Repository
}

func NewShortenerHandler(rep repositories.Repository) *ShortenerHandler {
	return &ShortenerHandler{
		Rep: rep,
	}
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

	id := h.Rep.ReserveID()
	shortURL := models.NewShortURL(id, url, models.UID(util.GenerateUniqueHash(id)))

	h.Rep.Save(shortURL)

	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(shortURL.GetShortURL()))

	if err != nil {
		log.Println(err.Error())
	}
}
