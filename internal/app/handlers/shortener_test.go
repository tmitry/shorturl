package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/handlers"
	"github.com/tmitry/shorturl/internal/app/middlewares"
	"github.com/tmitry/shorturl/internal/app/mocks"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

var errFailedToPing = errors.New("failed to ping")

func TestShortenerHandler_Shorten(t *testing.T) {
	t.Parallel()

	type fields struct {
		Cfg              *configs.Config
		Rep              repositories.Repository
		ContextKeyUserID middlewares.ContextKey
	}

	type request struct {
		body   string
		userID any
	}

	type response struct {
		statusCode  int
		contentType string
		body        string
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// test case 1
	cfg1 := configs.NewDefaultConfig()
	rep1 := mocks.NewMockRepository(ctrl)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	rep2 := mocks.NewMockRepository(ctrl)

	// test case 3
	cfg3 := configs.NewDefaultConfig()
	cfg3.App.HashMinLength = 3
	cfg3.App.HashSalt = "abc"
	cfg3.Server.BaseURL = "localhost"
	rep3 := mocks.NewMockRepository(ctrl)
	url3 := "https://example.com/"
	userID3 := uuid.New()
	shortURL3 := models.NewShortURL(1, models.URL(url3), "uid", userID3)
	rep3.EXPECT().Save(
		gomock.Any(),
		models.URL(url3),
		userID3,
		cfg3.App.HashMinLength,
		cfg3.App.HashSalt,
	).Return(shortURL3, nil)

	tests := []struct {
		name     string
		fields   fields
		request  request
		response response
	}{
		{
			name: "test case 1: incorrect user id",
			fields: fields{
				Cfg:              cfg1,
				Rep:              rep1,
				ContextKeyUserID: "userID",
			},
			request: request{
				body:   "https://example.com/",
				userID: "bad user id value",
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "test case 2: incorrect url",
			fields: fields{
				Cfg:              cfg2,
				Rep:              rep2,
				ContextKeyUserID: "userID",
			},
			request: request{
				body:   "bad url",
				userID: uuid.New(),
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				contentType: handlers.ContentTypeText,
				body: fmt.Sprintf(
					"%s: %s",
					http.StatusText(http.StatusBadRequest),
					handlers.MessageIncorrectURL,
				),
			},
		},
		{
			name: "test case 3: created",
			fields: fields{
				Cfg:              cfg3,
				Rep:              rep3,
				ContextKeyUserID: "userID",
			},
			request: request{
				body:   url3,
				userID: userID3,
			},
			response: response{
				statusCode:  http.StatusCreated,
				contentType: handlers.ContentTypeText,
				body:        string(shortURL3.GetShortURL(cfg3.Server.BaseURL)),
			},
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shortenerHandler := handlers.ShortenerHandler{
				Cfg:              testCase.fields.Cfg,
				Rep:              testCase.fields.Rep,
				ContextKeyUserID: testCase.fields.ContextKeyUserID,
			}

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.request.body))
			request = request.WithContext(context.WithValue(
				request.Context(),
				testCase.fields.ContextKeyUserID,
				testCase.request.userID,
			))

			recorder := httptest.NewRecorder()

			shortenerHandler.Shorten(recorder, request)

			result := recorder.Result()

			assert.Equal(t, testCase.response.contentType, handlers.GetContentType(result))

			assert.Equal(t, testCase.response.statusCode, result.StatusCode)

			body, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, testCase.response.body, strings.TrimSuffix(string(body), "\n"))
		})
	}
}

func TestShortenerHandler_Redirect(t *testing.T) {
	t.Parallel()

	type fields struct {
		Cfg *configs.Config
		Rep repositories.Repository
	}

	type request struct {
		uid string
	}

	type response struct {
		statusCode  int
		contentType string
		body        string
		location    string
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// test case 1
	cfg1 := configs.NewDefaultConfig()
	cfg1.App.HashMinLength = 5
	rep1 := mocks.NewMockRepository(ctrl)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	cfg2.App.HashMinLength = 6
	rep2 := mocks.NewMockRepository(ctrl)
	uid2 := models.UID("AbCdEF")
	rep2.EXPECT().FindOneByUID(gomock.Any(), uid2).Return(nil, repositories.ErrNotFound)

	// test case 3
	cfg3 := configs.NewDefaultConfig()
	cfg3.App.HashMinLength = 7
	rep3 := mocks.NewMockRepository(ctrl)
	uid3 := models.UID("AbCdEFg")
	url3 := "https://example.com/"
	shortURL3 := models.NewShortURL(1, models.URL(url3), uid3, uuid.New())
	rep3.EXPECT().FindOneByUID(gomock.Any(), uid3).Return(shortURL3, nil)

	tests := []struct {
		name     string
		fields   fields
		request  request
		response response
	}{
		{
			name: "test case 1: incorrect uid",
			fields: fields{
				Cfg: cfg1,
				Rep: rep1,
			},
			request: request{
				uid: "abc",
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				contentType: handlers.ContentTypeText,
				body: fmt.Sprintf(
					"%s: %s",
					http.StatusText(http.StatusBadRequest),
					handlers.MessageIncorrectUID,
				),
				location: "",
			},
		},
		{
			name: "test case 2: url not found",
			fields: fields{
				Cfg: cfg2,
				Rep: rep2,
			},
			request: request{
				uid: uid2.String(),
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				contentType: handlers.ContentTypeText,
				body: fmt.Sprintf(
					"%s: %s",
					http.StatusText(http.StatusBadRequest),
					handlers.MessageURLNotFound,
				),
				location: "",
			},
		},
		{
			name: "test case 3: redirect",
			fields: fields{
				Cfg: cfg3,
				Rep: rep3,
			},
			request: request{
				uid: uid3.String(),
			},
			response: response{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: handlers.ContentTypeText,
				body:        "",
				location:    url3,
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			handler := handlers.ShortenerHandler{
				Cfg:              testCase.fields.Cfg,
				Rep:              testCase.fields.Rep,
				ContextKeyUserID: "",
			}

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add(handlers.ParameterNameUID, testCase.request.uid)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeCtx))

			recorder := httptest.NewRecorder()
			handler.Redirect(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.response.statusCode, result.StatusCode)
			assert.Equal(t, testCase.response.contentType, handlers.GetContentType(result))

			body, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, testCase.response.body, strings.TrimSuffix(string(body), "\n"))
			assert.Equal(t, testCase.response.location, result.Header.Get("Location"))
		})
	}
}

func TestShortenerHandler_Ping(t *testing.T) {
	t.Parallel()

	type fields struct {
		Cfg *configs.Config
		Rep repositories.Repository
	}

	type response struct {
		statusCode  int
		contentType string
		body        string
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// test case 1
	cfg1 := configs.NewDefaultConfig()
	rep1 := mocks.NewMockRepository(ctrl)
	rep1.EXPECT().Ping(gomock.Any()).Return(errFailedToPing)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	rep2 := mocks.NewMockRepository(ctrl)
	rep2.EXPECT().Ping(gomock.Any()).Return(nil)

	tests := []struct {
		name     string
		fields   fields
		response response
	}{
		{
			name: "test case 1: failed to ping",
			fields: fields{
				Cfg: cfg1,
				Rep: rep1,
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "test case 2: ok",
			fields: fields{
				Cfg: cfg2,
				Rep: rep2,
			},
			response: response{
				statusCode:  http.StatusOK,
				contentType: handlers.ContentTypeText,
				body:        "",
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			handler := handlers.ShortenerHandler{
				Cfg:              testCase.fields.Cfg,
				Rep:              testCase.fields.Rep,
				ContextKeyUserID: "",
			}

			request := httptest.NewRequest(http.MethodGet, "/ping", nil)

			recorder := httptest.NewRecorder()
			handler.Ping(recorder, request)
			result := recorder.Result()

			assert.Equal(t, testCase.response.statusCode, result.StatusCode)
			assert.Equal(t, testCase.response.contentType, handlers.GetContentType(result))

			body, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, testCase.response.body, strings.TrimSuffix(string(body), "\n"))
		})
	}
}
