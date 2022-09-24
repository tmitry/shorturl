package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const (
	ParameterNameUID = "uid"
)

type ShortenerHandler struct {
	Cfg              *configs.Config
	Rep              repositories.Repository
	ContextKeyUserID middlewares.ContextKey
}

func NewShortenerHandler(
	cfg *configs.Config,
	rep repositories.Repository,
	contextKeyUserID middlewares.ContextKey,
) *ShortenerHandler {
	return &ShortenerHandler{
		Cfg:              cfg,
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
		http.Error(writer, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectURL),
			http.StatusBadRequest)

		return
	}

	shortURL, err := h.Rep.Save(request.Context(), url, userID, h.Cfg.App.HashMinLength, h.Cfg.App.HashSalt)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusCreated)

	_, err = writer.Write([]byte(shortURL.GetShortURL(h.Cfg.Server.BaseURL)))
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerHandler) Redirect(writer http.ResponseWriter, request *http.Request) {
	uid := models.UID(chi.URLParam(request, ParameterNameUID))

	isValid, err := uid.IsValid(h.Cfg.App.HashMinLength)
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

	shortURL, err := h.Rep.FindOneByUID(request.Context(), uid)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			http.Error(
				writer,
				fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageURLNotFound),
				http.StatusBadRequest,
			)

			return
		}

		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Location", shortURL.URL.String())
	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (h ShortenerHandler) Ping(writer http.ResponseWriter, request *http.Request) {
	if err := h.Rep.Ping(request.Context()); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusOK)
}
