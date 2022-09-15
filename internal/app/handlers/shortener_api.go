package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

type shortenRequestJSON struct {
	URL models.URL `json:"url"`
}

func NewShortenRequestJSON() *shortenRequestJSON {
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

func NewUserUrlsResponseJSON(userShortURLs []*models.ShortURL) interface{} {
	response := make([]struct {
		ShortURL    models.URL `json:"short_url"`
		OriginalURL models.URL `json:"original_url"`
	}, 0, len(userShortURLs))

	for _, userShortURL := range userShortURLs {
		response = append(response, struct {
			ShortURL    models.URL `json:"short_url"`
			OriginalURL models.URL `json:"original_url"`
		}{ShortURL: userShortURL.GetShortURL(), OriginalURL: userShortURL.URL})
	}

	return &response
}

type ShortenerAPIHandler struct {
	Rep              repositories.Repository
	ContextKeyUserID middlewares.ContextKey
}

func NewShortenerAPIHandler(rep repositories.Repository, contextKeyUserID middlewares.ContextKey) *ShortenerAPIHandler {
	return &ShortenerAPIHandler{
		Rep:              rep,
		ContextKeyUserID: contextKeyUserID,
	}
}

func (h ShortenerAPIHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
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

	requestJSON := NewShortenRequestJSON()
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

	id := h.Rep.ReserveID()
	shortURL := models.NewShortURL(id, requestJSON.URL, models.GenerateUID(id), userID)

	h.Rep.Save(shortURL)

	writer.Header().Set("Content-Type", ContentTypeJSON)
	writer.WriteHeader(http.StatusCreated)

	responseJSON := NewShortenResponseJSON(shortURL.GetShortURL())

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
	userID, ok := request.Context().Value(h.ContextKeyUserID).(uuid.UUID)
	if !ok {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(MessageIncorrectUserID)

		return
	}

	userShortURLs := h.Rep.FindAllByUserID(userID)
	if userShortURLs == nil {
		http.Error(writer, http.StatusText(http.StatusNoContent), http.StatusNoContent)

		return
	}

	writer.Header().Set("Content-Type", ContentTypeJSON)

	responseJSON := NewUserUrlsResponseJSON(userShortURLs)

	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetEscapeHTML(false)

	err := jsonEncoder.Encode(responseJSON)
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
