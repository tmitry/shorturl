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
)

func TestShortenerHandler_Shorten(t *testing.T) {
	t.Parallel()

	rep := repositories.NewMemoryRepository()
	handler := NewShortenerHandler(rep)

	type want struct {
		contentType string
		statusCode  int
		content     string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "bad URL",
			url:  "bad_url",
			want: want{
				contentType: ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			},
		},
		{
			name: "correct URL",
			url:  "https://example.com/",
			want: want{
				contentType: ContentTypeText,
				statusCode:  http.StatusCreated,
				content:     "",
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, config.ServerCfg.Address, strings.NewReader(testCase.url))

			recorder := httptest.NewRecorder()

			handler.Shorten(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.want.statusCode, result.StatusCode)
			assert.Equal(t, testCase.want.contentType, GetContentType(result))

			content, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if result.StatusCode != http.StatusCreated {
				assert.Equal(t, testCase.want.content, strings.TrimSuffix(string(content), "\n"))
			}
		})
	}
}

func TestShortenerHandler_Redirect(t *testing.T) {
	t.Parallel()

	rep := repositories.NewMemoryRepository()
	handler := NewShortenerHandler(rep)

	id := rep.ReserveID()
	url := models.URL("https://example.com/")
	shortURL := models.NewShortURL(id, url, models.GenerateUID(id))
	rep.Save(shortURL)

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
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "wrong UID parameter name",
			request: map[string]string{"u_id": "gIJsL"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "empty UID parameter",
			request: map[string]string{"uid": ""},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID has few characters",
			request: map[string]string{ParameterNameUID: "asd"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID has invalid characters",
			request: map[string]string{ParameterNameUID: "a#s-d!a"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectUID),
				location:    "",
			},
		},
		{
			name:    "UID not found",
			request: map[string]string{ParameterNameUID: "AJsGF"},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: ContentTypeText,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageURLNotFound),
				location:    "",
			},
		},
		{
			name:    "correct UID",
			request: map[string]string{ParameterNameUID: shortURL.UID.String()},
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: ContentTypeText,
				content:     "",
				location:    shortURL.URL.String(),
			},
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, config.ServerCfg.Address, nil)
			request.Header.Set("Content-Type", ContentTypeText)

			routeCtx := chi.NewRouteContext()
			for param, value := range testCase.request {
				routeCtx.URLParams.Add(param, value)
			}
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeCtx))

			recorder := httptest.NewRecorder()

			handler.Redirect(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.want.statusCode, result.StatusCode)
			assert.Equal(t, testCase.want.contentType, GetContentType(result))
			assert.Equal(t, testCase.want.location, result.Header.Get("Location"))

			content, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, testCase.want.content, strings.TrimSuffix(string(content), "\n"))
		})
	}
}

func GetContentType(r *http.Response) string {
	s := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if i := strings.Index(s, ";"); i > -1 {
		s = s[0:i]
	}

	return s
}
