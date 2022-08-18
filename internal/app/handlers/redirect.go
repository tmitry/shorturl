package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

const (
	ParameterNameUID = "uid"

	messageIncorrectUID = "incorrect UID"
	messageURLNotFound  = "URL not found"
)

type RedirectHandler struct {
	Rep repositories.Repository
}

func NewRedirectHandler(rep repositories.Repository) *RedirectHandler {
	return &RedirectHandler{
		Rep: rep,
	}
}

func (h RedirectHandler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if _, ok := vars[ParameterNameUID]; !ok {
		http.NotFound(writer, r)

		return
	}

	uid := models.UID(vars[ParameterNameUID])

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
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusTemporaryRedirect)
}
