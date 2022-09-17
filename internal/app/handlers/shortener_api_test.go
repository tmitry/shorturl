package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func TestShortenerAPIHandler_Shorten(t *testing.T) {
	t.Parallel()

	const (
		ContextKeyUserID middlewares.ContextKey = "userID"
	)

	cfg := configs.NewDefaultConfig()
	rep := repositories.NewMemoryRepository()
	handler := handlers.NewShortenerAPIHandler(cfg, rep, ContextKeyUserID)

	type want struct {
		contentType string
		statusCode  int
		content     string
	}

	tests := []struct {
		name string
		json string
		want want
	}{
		{
			name: "bad JSON",
			json: "bad_json",
			want: want{
				contentType: handlers.ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), handlers.MessageIncorrectJSON),
			},
		},
		{
			name: "bad JSON format",
			json: `{"ur_l":"https://example.com/"}`,
			want: want{
				contentType: handlers.ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), handlers.MessageIncorrectURL),
			},
		},
		{
			name: "bad URL",
			json: `{"url":"bad_url"}`,
			want: want{
				contentType: handlers.ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), handlers.MessageIncorrectURL),
			},
		},
		{
			name: "correct URL",
			json: `{"url":"https://example.com/"}`,
			want: want{
				contentType: handlers.ContentTypeJSON,
				statusCode:  http.StatusCreated,
				content:     "",
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, cfg.Server.Address, strings.NewReader(testCase.json))
			request.Header.Set("Content-Type", handlers.ContentTypeJSON)
			request = request.WithContext(context.WithValue(request.Context(), ContextKeyUserID, uuid.New()))

			recorder := httptest.NewRecorder()

			handler.Shorten(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.want.statusCode, result.StatusCode)
			assert.Equal(t, testCase.want.contentType, handlers.GetContentType(result))

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

func TestShortenerAPIHandler_UserUrls(t *testing.T) {
	t.Parallel()

	const (
		ContextKeyUserID middlewares.ContextKey = "userID"
	)

	cfg := configs.NewDefaultConfig()
	rep := repositories.NewMemoryRepository()
	handler := handlers.NewShortenerAPIHandler(cfg, rep, ContextKeyUserID)

	id := rep.ReserveID()
	url := models.URL("https://example.com/")
	userID := uuid.New()
	shortURL := models.NewShortURL(id, url, models.NewUID(id, cfg.App.HashMinLength, cfg.App.HashSalt), userID)
	rep.Save(shortURL)

	content, err := json.Marshal(handlers.NewUserUrlsResponseJSON([]*models.ShortURL{shortURL}, cfg.Server.BaseURL))
	assert.NoError(t, err)

	type want struct {
		contentType string
		statusCode  int
		content     string
	}

	tests := []struct {
		name   string
		userID uuid.UUID
		want   want
	}{
		{
			name:   "urls not exist",
			userID: uuid.New(),
			want: want{
				contentType: handlers.ContentTypeText,
				statusCode:  http.StatusNoContent,
				content:     http.StatusText(http.StatusNoContent),
			},
		},
		{
			name:   "urls exist",
			userID: userID,
			want: want{
				contentType: handlers.ContentTypeJSON,
				statusCode:  http.StatusOK,
				content:     string(content),
			},
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, cfg.Server.Address, nil)
			request = request.WithContext(context.WithValue(request.Context(), ContextKeyUserID, testCase.userID))

			recorder := httptest.NewRecorder()

			handler.UserUrls(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.want.statusCode, result.StatusCode)
			assert.Equal(t, testCase.want.contentType, handlers.GetContentType(result))

			content, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if testCase.want.contentType == handlers.ContentTypeJSON {
				assert.JSONEq(t, testCase.want.content, string(content))
			} else {
				assert.Equal(t, testCase.want.content, strings.TrimSuffix(string(content), "\n"))
			}
		})
	}
}
