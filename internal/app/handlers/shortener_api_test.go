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

func TestShortenerAPIHandler_Shorten(t *testing.T) {
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
	rep3 := mocks.NewMockRepository(ctrl)

	// test case 4
	cfg4 := configs.NewDefaultConfig()
	cfg4.Server.BaseURL = "baseurl"
	uid4 := models.UID("AbCdEf")
	uidGenerator4 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator4.EXPECT().Generate().Return(uid4, nil)

	rep4 := mocks.NewMockRepository(ctrl)
	url4 := "https://example-site.com/"
	body4 := fmt.Sprintf(`{"url":"%s"}`, url4)
	userID4 := uuid.New()
	shortURL4 := models.NewShortURL(0, models.URL(url4), uid4, userID4)
	rep4.EXPECT().Save(gomock.Any(), shortURL4).Return(nil)
	json4, err := json.Marshal(handlers.NewShortenResponseJSON(shortURL4.GetShortURL(cfg4.Server.BaseURL)))
	require.NoError(t, err)

	// test case 5
	cfg5 := configs.NewDefaultConfig()
	cfg5.Server.BaseURL = "base-url"
	uid5 := models.UID("SdfGfs")
	uidGenerator5 := mocks.NewMockUIDGenerator(ctrl)
	uidGenerator5.EXPECT().Generate().Return(uid5, nil)

	rep5 := mocks.NewMockRepository(ctrl)
	url5 := "https://example-site.com/conflict"
	body5 := fmt.Sprintf(`{"url":"%s"}`, url5)
	userID5 := uuid.New()
	shortURL5 := models.NewShortURL(0, models.URL(url5), uid5, userID5)
	rep5.EXPECT().Save(gomock.Any(), shortURL5).Return(repositories.ErrURLDuplicate)
	json5, err := json.Marshal(handlers.NewShortenResponseJSON(shortURL5.GetShortURL(cfg5.Server.BaseURL)))
	require.NoError(t, err)

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
				contextKeyUserID: "userId",
			},
			request: request{
				body:   `{"url":"https://example1.com/"}`,
				userID: "bad user id value",
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "test case 2: incorrect json",
			fields: fields{
				cfg:              cfg2,
				uidGenerator:     uidGenerator2,
				rep:              rep2,
				contextKeyUserID: "userId",
			},
			request: request{
				body:   `bad json`,
				userID: uuid.New(),
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				contentType: handlers.ContentTypeText,
				body: fmt.Sprintf(
					"%s: %s",
					http.StatusText(http.StatusBadRequest),
					handlers.MessageIncorrectJSON,
				),
			},
		},
		{
			name: "test case 3: incorrect url",
			fields: fields{
				cfg:              cfg3,
				uidGenerator:     uidGenerator3,
				rep:              rep3,
				contextKeyUserID: "userId",
			},
			request: request{
				body:   `{"url":"incorrect url"}`,
				userID: uuid.New(),
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				contentType: handlers.ContentTypeText,
				body:        fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), handlers.MessageIncorrectURL),
			},
		},
		{
			name: "test case 4: created",
			fields: fields{
				cfg:              cfg4,
				uidGenerator:     uidGenerator4,
				rep:              rep4,
				contextKeyUserID: "userId",
			},
			request: request{
				body:   body4,
				userID: userID4,
			},
			response: response{
				statusCode:  http.StatusCreated,
				contentType: handlers.ContentTypeJSON,
				body:        string(json4),
			},
		},
		{
			name: "test case 5: conflict",
			fields: fields{
				cfg:              cfg5,
				uidGenerator:     uidGenerator5,
				rep:              rep5,
				contextKeyUserID: "USERID",
			},
			request: request{
				body:   body5,
				userID: userID5,
			},
			response: response{
				statusCode:  http.StatusConflict,
				contentType: handlers.ContentTypeJSON,
				body:        string(json5),
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shortenerAPIHandler := handlers.NewShortenerAPIHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				testCase.fields.contextKeyUserID,
			)

			requestAPIShorten := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(testCase.request.body))
			requestAPIShorten = requestAPIShorten.WithContext(context.WithValue(
				requestAPIShorten.Context(),
				testCase.fields.contextKeyUserID,
				testCase.request.userID,
			))

			recorder := httptest.NewRecorder()

			shortenerAPIHandler.Shorten(recorder, requestAPIShorten)
			result := recorder.Result()

			body, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, testCase.response.body, strings.TrimSuffix(string(body), "\n"))
			assert.Equal(t, testCase.response.contentType, handlers.GetContentType(result))
			assert.Equal(t, testCase.response.statusCode, result.StatusCode)
		})
	}
}

func TestShortenerAPIHandler_UserUrls(t *testing.T) {
	t.Parallel()

	type fields struct {
		cfg              *configs.Config
		uidGenerator     utils.UIDGenerator
		rep              repositories.Repository
		contextKeyUserID middlewares.ContextKey
	}

	type request struct {
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
	userID2 := uuid.New()
	rep2 := mocks.NewMockRepository(ctrl)
	rep2.EXPECT().FindAllByUserID(gomock.Any(), userID2).Return(nil, repositories.ErrNotFound)

	// test case 3
	cfg3 := configs.NewDefaultConfig()
	cfg3.Server.BaseURL = "host"
	uidGenerator3 := mocks.NewMockUIDGenerator(ctrl)
	userID3 := uuid.New()
	rep3 := mocks.NewMockRepository(ctrl)

	var userShortURLs []*models.ShortURL

	userShortURLs = append(
		userShortURLs,
		models.NewShortURL(1, models.URL("https://example.com/1"), "uid1", userID3),
	)
	userShortURLs = append(
		userShortURLs,
		models.NewShortURL(2, models.URL("https://example.com/2"), "uid2", userID3),
	)
	rep3.EXPECT().FindAllByUserID(gomock.Any(), userID3).Return(userShortURLs, nil)
	json3, err := json.Marshal(handlers.NewUserUrlsResponseJSON(userShortURLs, cfg3.Server.BaseURL))
	require.NoError(t, err)

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
				userID: "bad user id value",
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "test case 2: urls not found",
			fields: fields{
				cfg:              cfg2,
				uidGenerator:     uidGenerator2,
				rep:              rep2,
				contextKeyUserID: "userID",
			},
			request: request{
				userID: userID2,
			},
			response: response{
				statusCode:  http.StatusNoContent,
				contentType: handlers.ContentTypeText,
				body:        http.StatusText(http.StatusNoContent),
			},
		},
		{
			name: "test case 3: urls found",
			fields: fields{
				cfg:              cfg3,
				uidGenerator:     uidGenerator3,
				rep:              rep3,
				contextKeyUserID: "userID",
			},
			request: request{
				userID: userID3,
			},
			response: response{
				statusCode:  http.StatusOK,
				contentType: handlers.ContentTypeJSON,
				body:        string(json3),
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shortenerAPIHandler := handlers.NewShortenerAPIHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				testCase.fields.contextKeyUserID,
			)

			request := httptest.NewRequest(http.MethodPost, "/api/user/urls", nil)
			request = request.WithContext(context.WithValue(
				request.Context(),
				testCase.fields.contextKeyUserID,
				testCase.request.userID,
			))

			recorder := httptest.NewRecorder()
			shortenerAPIHandler.UserUrls(recorder, request)
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

func TestShortenerAPIHandler_ShortenBatch(t *testing.T) {
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
		body        string
		contentType string
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
	rep3 := mocks.NewMockRepository(ctrl)

	// test case 4
	cfg4 := configs.NewDefaultConfig()
	cfg4.Server.BaseURL = "localhost"
	uid41 := models.UID("AbCdEf")
	uid42 := models.UID("DcDsGd")
	uidGenerator4 := mocks.NewMockUIDGenerator(ctrl)
	generate1 := uidGenerator4.EXPECT().Generate().Return(uid41, nil)
	generate2 := uidGenerator4.EXPECT().Generate().Return(uid42, nil)
	gomock.InOrder(generate1, generate2)

	rep4 := mocks.NewMockRepository(ctrl)
	urls := []models.URL{"https://mysite.com?id=u1", "https://mysite.com?id=u2"}
	correlationIDs := []string{"u1", "u2"}
	body4 := fmt.Sprintf(`[
    {
        "correlation_id": "%s",
        "original_url": "%s"
    },
    {
        "correlation_id": "%s",
        "original_url": "%s"
    }
]`, correlationIDs[0], urls[0], correlationIDs[1], urls[1])
	userID4 := uuid.New()
	shortURLs := []*models.ShortURL{
		models.NewShortURL(0, urls[0], uid41, userID4),
		models.NewShortURL(0, urls[1], uid42, userID4),
	}
	rep4.EXPECT().BatchSave(gomock.Any(), shortURLs).Return(nil)
	json4, err := json.Marshal(handlers.NewShortenBatchResponseJSON(
		shortURLs,
		correlationIDs,
		cfg4.Server.BaseURL),
	)
	require.NoError(t, err)

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
				body:   `[{"correlation_id": "u1", "original_url": "https://mysite.com?id=u1"}]`,
				userID: "bad user id value",
			},
			response: response{
				statusCode:  http.StatusInternalServerError,
				body:        http.StatusText(http.StatusInternalServerError),
				contentType: handlers.ContentTypeText,
			},
		},
		{
			name: "test case 2: incorrect json",
			fields: fields{
				cfg:              cfg2,
				uidGenerator:     uidGenerator2,
				rep:              rep2,
				contextKeyUserID: "userID",
			},
			request: request{
				body:   `bad json`,
				userID: uuid.New(),
			},
			response: response{
				statusCode: http.StatusBadRequest,
				body: fmt.Sprintf(
					"%s: %s",
					http.StatusText(http.StatusBadRequest),
					handlers.MessageIncorrectJSON,
				),
				contentType: handlers.ContentTypeText,
			},
		},
		{
			name: "test case 3: incorrect url",
			fields: fields{
				cfg:              cfg3,
				uidGenerator:     uidGenerator3,
				rep:              rep3,
				contextKeyUserID: "userID",
			},
			request: request{
				body:   `[{"correlation_id": "u1", "original_url": "bad url"}]`,
				userID: uuid.New(),
			},
			response: response{
				statusCode:  http.StatusBadRequest,
				body:        fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), handlers.MessageIncorrectURL),
				contentType: handlers.ContentTypeText,
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
				body:   body4,
				userID: userID4,
			},
			response: response{
				statusCode:  http.StatusCreated,
				body:        string(json4),
				contentType: handlers.ContentTypeJSON,
			},
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			shortenerAPIHandler := handlers.NewShortenerAPIHandler(
				testCase.fields.cfg,
				testCase.fields.uidGenerator,
				testCase.fields.rep,
				testCase.fields.contextKeyUserID,
			)

			requestShortenAPIBatch := httptest.NewRequest(
				http.MethodPost,
				"/api/shorten/batch",
				strings.NewReader(testCase.request.body),
			)
			requestShortenAPIBatch = requestShortenAPIBatch.WithContext(context.WithValue(
				requestShortenAPIBatch.Context(),
				testCase.fields.contextKeyUserID,
				testCase.request.userID,
			))

			recorder := httptest.NewRecorder()
			shortenerAPIHandler.ShortenBatch(recorder, requestShortenAPIBatch)
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
