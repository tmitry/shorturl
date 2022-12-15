package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
	"github.com/tmitry/shorturl/internal/app/utils"
)

type shortenRequestJSON struct {
	URL models.URL `json:"url"`
}

func newShortenRequestJSON() *shortenRequestJSON {
	return &shortenRequestJSON{
		URL: "",
	}
}

func NewShortenResponseJSON(url models.URL) interface{} {
	response := struct {
		Result models.URL `json:"result"`
	}{Result: url}

	return &response
}

func NewUserUrlsResponseJSON(userShortURLs []*models.ShortURL, baseURL string) interface{} {
	response := make([]struct {
		ShortURL    models.URL `json:"short_url"`
		OriginalURL models.URL `json:"original_url"`
	}, 0, len(userShortURLs))

	for _, userShortURL := range userShortURLs {
		response = append(response, struct {
			ShortURL    models.URL `json:"short_url"`
			OriginalURL models.URL `json:"original_url"`
		}{ShortURL: userShortURL.GetShortURL(baseURL), OriginalURL: userShortURL.URL})
	}

	return &response
}

type shortenBatchItemRequestJSON struct {
	CorrelationID string     `json:"correlation_id"`
	OriginalURL   models.URL `json:"original_url"`
}

type ShortenBatchRequestJSON []shortenBatchItemRequestJSON

func NewShortenBatchResponseJSON(shortURLs []*models.ShortURL, correlationIDs []string, baseURL string) interface{} {
	response := make([]struct {
		CorrelationID string     `json:"correlation_id"`
		ShortURL      models.URL `json:"short_url"`
	}, 0, len(shortURLs))

	for index, userShortURL := range shortURLs {
		response = append(response, struct {
			CorrelationID string     `json:"correlation_id"`
			ShortURL      models.URL `json:"short_url"`
		}{CorrelationID: correlationIDs[index], ShortURL: userShortURL.GetShortURL(baseURL)})
	}

	return &response
}

type ShortenerAPIHandler struct {
	cfg              *configs.Config
	uidGenerator     utils.UIDGenerator
	rep              repositories.Repository
	contextKeyUserID middlewares.ContextKey
	deletionBuffer   utils.DeletionBuffer
}

func NewShortenerAPIHandler(
	cfg *configs.Config,
	uidGenerator utils.UIDGenerator,
	rep repositories.Repository,
	contextKeyUserID middlewares.ContextKey,
	deletionBuffer utils.DeletionBuffer,
) *ShortenerAPIHandler {
	return &ShortenerAPIHandler{
		cfg:              cfg,
		uidGenerator:     uidGenerator,
		rep:              rep,
		contextKeyUserID: contextKeyUserID,
		deletionBuffer:   deletionBuffer,
	}
}

func (h ShortenerAPIHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
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

	requestJSON := newShortenRequestJSON()
	if err := json.NewDecoder(reader).Decode(requestJSON); err != nil {
		http.Error(writer, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectJSON),
			http.StatusBadRequest)

		return
	}

	if !requestJSON.URL.IsValid() {
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

	shortURL := models.NewShortURL(0, requestJSON.URL, uid, userID)

	if err := h.rep.Save(request.Context(), shortURL); err != nil {
		if !errors.Is(err, repositories.ErrURLDuplicate) {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())

			return
		}

		statusCode = http.StatusConflict
	}

	writer.Header().Set("Content-Type", ContentTypeJSON)
	writer.WriteHeader(statusCode)

	responseJSON := NewShortenResponseJSON(shortURL.GetShortURL(h.cfg.Server.BaseURL))

	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetEscapeHTML(false)

	err = jsonEncoder.Encode(responseJSON)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	_, err = buf.WriteTo(writer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerAPIHandler) UserUrls(writer http.ResponseWriter, request *http.Request) {
	userID, ok := request.Context().Value(h.contextKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(MessageIncorrectUserID)

		return
	}

	userShortURLs, err := h.rep.FindAllByUserID(request.Context(), userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			http.Error(writer, http.StatusText(http.StatusNoContent), http.StatusNoContent)

			return
		}

		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Content-Type", ContentTypeJSON)

	responseJSON := NewUserUrlsResponseJSON(userShortURLs, h.cfg.Server.BaseURL)

	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetEscapeHTML(false)

	err = jsonEncoder.Encode(responseJSON)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	_, err = buf.WriteTo(writer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerAPIHandler) ShortenBatch(writer http.ResponseWriter, request *http.Request) {
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

	var requestJSON ShortenBatchRequestJSON

	if err := json.NewDecoder(reader).Decode(&requestJSON); err != nil {
		http.Error(writer, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectJSON),
			http.StatusBadRequest)

		return
	}

	shortURLs := make([]*models.ShortURL, 0, len(requestJSON))
	correlationIDs := make([]string, 0, len(requestJSON))

	for _, item := range requestJSON {
		if !item.OriginalURL.IsValid() {
			http.Error(writer, fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectURL),
				http.StatusBadRequest)

			return
		}

		uid, err := h.uidGenerator.Generate()
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())

			return
		}

		shortURL := models.NewShortURL(0, item.OriginalURL, uid, userID)

		shortURLs = append(shortURLs, shortURL)
		correlationIDs = append(correlationIDs, item.CorrelationID)
	}

	err = h.rep.BatchSave(request.Context(), shortURLs)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	writer.Header().Set("Content-Type", ContentTypeJSON)
	writer.WriteHeader(http.StatusCreated)

	responseJSON := NewShortenBatchResponseJSON(shortURLs, correlationIDs, h.cfg.Server.BaseURL)

	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetEscapeHTML(false)

	err = jsonEncoder.Encode(responseJSON)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}

	_, err = buf.WriteTo(writer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())

		return
	}
}

func (h ShortenerAPIHandler) DeleteUserUrls(writer http.ResponseWriter, request *http.Request) {
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

	var uids []models.UID

	if err := json.NewDecoder(reader).Decode(&uids); err != nil {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), MessageIncorrectJSON),
			http.StatusBadRequest,
		)

		return
	}

	h.deletionBuffer.Push(uids, userID)

	writer.Header().Set("Content-Type", ContentTypeText)
	writer.WriteHeader(http.StatusAccepted)
}
