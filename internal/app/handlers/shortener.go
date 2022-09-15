package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const (
	ParameterNameUID = "uid"
)

type ShortenerHandler struct {
	Rep              repositories.Repository
	ContextKeyUserID middlewares.ContextKey
}

func NewShortenerHandler(rep repositories.Repository, contextKeyUserID middlewares.ContextKey) *ShortenerHandler {
	return &ShortenerHandler{
		Rep:              rep,
		ContextKeyUserID: contextKeyUserID,
	}
}

func (h ShortenerHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(h.ContextKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(MessageIncorrectUserID)

		return
	}

	reader, err := getRequestReader(request)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())
		}
	}(reader)

	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	url := models.URL(body)

	if !url.IsValid() {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectURL),
			http.StatusBadRequest,
		)

		return
	}

	id := h.Rep.ReserveID()
	shortURL := models.NewShortURL(id, url, models.GenerateUID(id), userID)

	h.Rep.Save(shortURL)

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusCreated)

	_, err = writer.Write([]byte(shortURL.GetShortURL()))
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerHandler) Redirect(writer http.ResponseWriter, request *http.Request) {
	uid := models.UID(chi.URLParam(request, ParameterNameUID))

	isValid, err := uid.IsValid()
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	if !isValid {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectUID),
			http.StatusBadRequest,
		)

		return
	}

	shortURL := h.Rep.FindOneByUID(uid)
	if shortURL == nil {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageURLNotFound),
			http.StatusBadRequest,
		)

		return
	}

	writer.Header().Set("Location", shortURL.URL.String())
	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}
