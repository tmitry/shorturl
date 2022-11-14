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
	"github.com/tmitry/shorturl/internal/app/utils"
)

const (
	ParameterNameUID = "uid"
)

type ShortenerHandler struct {
	cfg              *configs.Config
	uidGenerator     utils.UIDGenerator
	rep              repositories.Repository
	contextKeyUserID middlewares.ContextKey
}

func NewShortenerHandler(
	cfg *configs.Config,
	uidGenerator utils.UIDGenerator,
	rep repositories.Repository,
	contextKeyUserID middlewares.ContextKey,
) *ShortenerHandler {
	return &ShortenerHandler{
		cfg:              cfg,
		uidGenerator:     uidGenerator,
		rep:              rep,
		contextKeyUserID: contextKeyUserID,
	}
}

func (h ShortenerHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(h.contextKeyUserID).(uuid.UUID)
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

	statusCode := http.StatusCreated

	uid, err := h.uidGenerator.Generate()
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	shortURL := models.NewShortURL(0, url, uid, userID)

	err = h.rep.Save(request.Context(), shortURL)
	if err != nil {
		if !errors.Is(err, repositories.ErrURLDuplicate) {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())

			return
		}

		statusCode = http.StatusConflict
	}

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(statusCode)

	_, err = writer.Write([]byte(shortURL.GetShortURL(h.cfg.Server.BaseURL)))
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerHandler) Redirect(writer http.ResponseWriter, request *http.Request) {
	uid := models.UID(chi.URLParam(request, ParameterNameUID))

	isValid, err := h.uidGenerator.IsValid(uid)
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

	shortURL, err := h.rep.FindOneByUID(request.Context(), uid)
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

	if shortURL.IsDeleted {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusGone), MessageURLWasDeleted),
			http.StatusGone,
		)

		return
	}

	writer.Header().Set("Location", shortURL.URL.String())
	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (h ShortenerHandler) Ping(writer http.ResponseWriter, request *http.Request) {
	if err := h.rep.Ping(request.Context()); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusOK)
}
