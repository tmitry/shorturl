package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

type requestJSON struct {
	URL models.URL `json:"url"`
}

func NewRequestJSON() *requestJSON {
	return &requestJSON{
		URL: "",
	}
}

type responseJSON struct {
	Result models.URL `json:"result"`
}

func NewResponseJSON(url models.URL) *responseJSON {
	return &responseJSON{
		Result: url,
	}
}

type ShortenerAPIHandler struct {
	Rep repositories.Repository
}

func NewShortenerAPIHandler(rep repositories.Repository) *ShortenerAPIHandler {
	return &ShortenerAPIHandler{
		Rep: rep,
	}
}

func (h ShortenerAPIHandler) Shorten(writer http.ResponseWriter, request *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())
		}
	}(request.Body)

	requestJSON := NewRequestJSON()
	if err := json.NewDecoder(request.Body).Decode(requestJSON); err != nil {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectJSON),
			http.StatusBadRequest,
		)

		return
	}

	if !requestJSON.URL.IsValid() {
		http.Error(
			writer,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			http.StatusBadRequest,
		)

		return
	}

	id := h.Rep.ReserveID()
	shortURL := models.NewShortURL(id, requestJSON.URL, models.GenerateUID(id))

	h.Rep.Save(shortURL)

	writer.Header().Set("Content-Type", ContentTypeJSON)
	writer.WriteHeader(http.StatusCreated)

	responseJSON := NewResponseJSON(shortURL.GetShortURL())

	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetEscapeHTML(false)

	err := jsonEncoder.Encode(responseJSON)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
	}

	_, err = buf.WriteTo(writer)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}
