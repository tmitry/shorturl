package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const (
	ParameterNameUID = "uid"
)

type ShortenerHandler struct {
	Rep repositories.Repository
}

func NewShortenerHandler(rep repositories.Repository) *ShortenerHandler {
	return &ShortenerHandler{
		Rep: rep,
	}
}

func (h ShortenerHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	err = request.Body.Close()
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
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
	shortURL := models.NewShortURL(id, url, models.GenerateUID(id))

	h.Rep.Save(shortURL)

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusCreated)

	_, err = writer.Write([]byte(shortURL.GetShortURL()))
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}

func (h ShortenerHandler) Redirect(writer http.ResponseWriter, r *http.Request) {
	uid := models.UID(chi.URLParam(r, ParameterNameUID))

	isValid, err := uid.IsValid()
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	if !isValid {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
			http.StatusBadRequest,
		)

		return
	}

	shortURL := h.Rep.Find(uid)
	if shortURL == nil {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageURLNotFound),
			http.StatusBadRequest,
		)

		return
	}

	writer.Header().Set("Location", shortURL.URL.String())
	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}
