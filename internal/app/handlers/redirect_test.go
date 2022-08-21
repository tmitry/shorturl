package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
	"github.com/tmitry/shorturl/internal/app/util"
)

func TestRedirectHandler(t *testing.T) {
	rep := repositories.NewMemoryRepository()
	handler := http.Handler(NewRedirectHandler(rep))

	id := rep.ReserveID()
	url := models.URL("https://example.com/")
	shortURL := models.NewShortURL(id, url, models.UID(util.GenerateUniqueHash(id)))
	rep.Add(shortURL)

	type want struct {
		contentType string
		statusCode  int
		content     string
		location    string
	}

	tests := []struct {
		name    string
		request map[string]string
		want    want
	}{
		{
			name:    "UID parameter doesnt exist",
			request: map[string]string{},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "wrong UID parameter name",
			request: map[string]string{"u_id": "gIJsL"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "empty UID parameter",
			request: map[string]string{"uid": ""},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID has few characters",
			request: map[string]string{ParameterNameUID: "asd"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID has invalid characters",
			request: map[string]string{ParameterNameUID: "a#s-d!a"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID not found",
			request: map[string]string{ParameterNameUID: "AJsGF"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageURLNotFound),
				location:    "",
			},
		},
		{
			name:    "correct UID",
			request: map[string]string{ParameterNameUID: shortURL.UID.String()},
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain; charset=utf-8",
				content:     "",
				location:    shortURL.URL.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, config.DefaultAddr, nil)

			routeCtx := chi.NewRouteContext()
			for param, value := range tt.request {
				routeCtx.URLParams.Add(param, value)
			}
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeCtx))

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)
			result := recorder.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))

			content, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.content, strings.TrimSuffix(string(content), "\n"))
		})
	}
}
