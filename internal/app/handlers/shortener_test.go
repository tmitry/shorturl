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
	"github.com/tmitry/shorturl/internal/app/utils"
)

var (
	errFailedToPing     = errors.New("failed to ping")
	errFailedToGenerate = errors.New("failed to generate")
)

func TestShortenerHandler_Shorten(t *testing.T) {
	t.Parallel()

	type fields struct {
		cfg              *configs.Config
		uidGenerator     utils.UIDGenerator
		rep              repositories.Repository
		contextKeyUserID middlewares.ContextKey
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
	uidGenerator1 := mocks.NewMockUIDGenerator(ctrl)
	rep1 := mocks.NewMockRepository(ctrl)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	uidGenerator2 := mocks.NewMockUIDGenerator(ctrl)
	rep2 := mocks.NewMockRepository(ctrl)

	// test case 3
	cfg3 := configs.NewDefaultConfig()
	uidGenerator3 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator3.EXPECT().Generate().Return(
		models.UID(""),
		errFailedToGenerate,
	)

	rep3 := mocks.NewMockRepository(ctrl)

	// test case 4
	cfg4 := configs.NewDefaultConfig()
	cfg4.Server.BaseURL = "localhost"
	uid4 := models.UID("AzFsx")
	uidGenerator4 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator4.EXPECT().Generate().Return(uid4, nil)

	rep4 := mocks.NewMockRepository(ctrl)
	url4 := "https://example.com/"
	userID4 := uuid.New()
	shortURL4 := models.NewShortURL(0, models.URL(url4), uid4, userID4)
	rep4.EXPECT().Save(
		gomock.Any(),
		shortURL4,
	).Return(nil)

	// test case 5
	cfg5 := configs.NewDefaultConfig()
	cfg5.Server.BaseURL = "localhost-url"
	uid5 := models.UID("AzFsx")
	uidGenerator5 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator5.EXPECT().Generate().Return(uid4, nil)

	rep5 := mocks.NewMockRepository(ctrl)
	url5 := "https://example.com/conflict"
	userID5 := uuid.New()
	shortURL5 := models.NewShortURL(0, models.URL(url5), uid5, userID5)
	rep5.EXPECT().Save(
		gomock.Any(),
		shortURL5,
	).Return(repositories.ErrURLDuplicate)

	tests := []struct {
		name     string
		fields   fields
		request  request
		response response
	}{
		{
			name: "test case 1: incorrect user id",
			fields: fields{
				cfg:              cfg1,
				uidGenerator:     uidGenerator1,
				rep:              rep1,
				contextKeyUserID: "userID",
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
				cfg:              cfg2,
				uidGenerator:     uidGenerator2,
				rep:              rep2,
				contextKeyUserID: "userID",
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
			name: "test case 3: error generate uid",
			fields: fields{
				cfg:              cfg3,
				uidGenerator:     uidGenerator3,
				rep:              rep3,
				contextKeyUserID: "userID",
			},
			request: request{
				body:   "https://example.com/",
				userID: uuid.New(),
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "test case 4: created",
			fields: fields{
				cfg:              cfg4,
				uidGenerator:     uidGenerator4,
				rep:              rep4,
				contextKeyUserID: "userID",
			},
			request: request{
				body:   url4,
				userID: userID4,
			},
			response: response{
				statusCode:  http.StatusCreated,
				contentType: handlers.ContentTypeText,
				body:        string(shortURL4.GetShortURL(cfg4.Server.BaseURL)),
			},
		},
		{
			name: "test case 5: conflict",
			fields: fields{
				cfg:              cfg5,
				uidGenerator:     uidGenerator5,
				rep:              rep5,
				contextKeyUserID: "userID",
			},
			request: request{
				body:   url5,
				userID: userID5,
			},
			response: response{
				statusCode:  http.StatusConflict,
				contentType: handlers.ContentTypeText,
				body:        string(shortURL5.GetShortURL(cfg5.Server.BaseURL)),
			},
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shortenerHandler := handlers.NewShortenerHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				testCase.fields.contextKeyUserID,
			)

			requestShorten := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.request.body))
			requestShorten = requestShorten.WithContext(context.WithValue(
				requestShorten.Context(),
				testCase.fields.contextKeyUserID,
				testCase.request.userID,
			))

			recorderShorten := httptest.NewRecorder()

			shortenerHandler.Shorten(recorderShorten, requestShorten)

			result := recorderShorten.Result()

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
		cfg              *configs.Config
		uidGenerator     utils.UIDGenerator
		rep              repositories.Repository
		contextKeyUserID middlewares.ContextKey
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
	uid1 := models.UID("abc")
	uidGenerator1 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator1.EXPECT().IsValid(uid1).Return(false, nil)

	rep1 := mocks.NewMockRepository(ctrl)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	uid2 := models.UID("AbCdEF")
	uidGenerator2 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator2.EXPECT().IsValid(uid2).Return(true, nil)

	rep2 := mocks.NewMockRepository(ctrl)
	rep2.EXPECT().FindOneByUID(gomock.Any(), uid2).Return(nil, repositories.ErrNotFound)

	// test case 3
	cfg3 := configs.NewDefaultConfig()
	uid3 := models.UID("AbCdEFg")
	uidGenerator3 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator3.EXPECT().IsValid(uid3).Return(true, nil)

	rep3 := mocks.NewMockRepository(ctrl)
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
				cfg:              cfg1,
				uidGenerator:     uidGenerator1,
				rep:              rep1,
				contextKeyUserID: "userID",
			},
			request: request{
				uid: uid1.String(),
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
				cfg:              cfg2,
				uidGenerator:     uidGenerator2,
				rep:              rep2,
				contextKeyUserID: "userID",
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
				cfg:              cfg3,
				uidGenerator:     uidGenerator3,
				rep:              rep3,
				contextKeyUserID: "userID",
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

			handler := handlers.NewShortenerHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				testCase.fields.contextKeyUserID,
			)

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
		cfg          *configs.Config
		uidGenerator utils.UIDGenerator
		rep          repositories.Repository
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
	uidGenerator1 := mocks.NewMockUIDGenerator(ctrl)
	rep1 := mocks.NewMockRepository(ctrl)
	rep1.EXPECT().Ping(gomock.Any()).Return(errFailedToPing)

	// test case 2
	cfg2 := configs.NewDefaultConfig()
	uidGenerator2 := mocks.NewMockUIDGenerator(ctrl)
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
				cfg:          cfg1,
				uidGenerator: uidGenerator1,
				rep:          rep1,
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
				cfg:          cfg2,
				uidGenerator: uidGenerator2,
				rep:          rep2,
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

			handler := handlers.NewShortenerHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				"",
			)

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
