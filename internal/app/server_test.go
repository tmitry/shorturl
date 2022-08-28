package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/handlers"
)

func TestRouter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		path        string
		statusCode  int
		contentType string
	}{
		{
			name:        "not allowed method PUT",
			method:      resty.MethodPut,
			path:        "/",
			statusCode:  http.StatusMethodNotAllowed,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "not allowed method PATCH",
			method:      resty.MethodPatch,
			path:        "/",
			statusCode:  http.StatusMethodNotAllowed,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "empty GET path",
			method:      resty.MethodGet,
			path:        "/",
			statusCode:  http.StatusMethodNotAllowed,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "short GET path",
			method:      resty.MethodGet,
			path:        "/dsa",
			statusCode:  http.StatusNotFound,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "incorrect characters in GET path",
			method:      resty.MethodGet,
			path:        "/d$!sa",
			statusCode:  http.StatusNotFound,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "correct GET path",
			method:      resty.MethodGet,
			path:        "/J9KQP",
			statusCode:  http.StatusBadRequest,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "incorrect POST path",
			method:      resty.MethodPost,
			path:        "/ds",
			statusCode:  http.StatusNotFound,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "incorrect Content-Type for POST path",
			method:      resty.MethodPost,
			path:        "/",
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "correct POST path",
			method:      resty.MethodPost,
			path:        "/",
			statusCode:  http.StatusBadRequest,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "not allowed method GET for API",
			method:      resty.MethodPut,
			path:        "/api/shorten",
			statusCode:  http.StatusMethodNotAllowed,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "incorrect Content-Type for API",
			method:      resty.MethodPost,
			path:        "/api/shorten",
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "correct API",
			method:      resty.MethodPost,
			path:        "/api/shorten",
			statusCode:  http.StatusBadRequest,
			contentType: handlers.ContentTypeJSON,
		},
	}

	router := NewRouter()
	testServer := httptest.NewServer(router)

	t.Cleanup(func() {
		testServer.Close()
	})

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resp := testRequest(t, testServer, testCase.method, testCase.path, testCase.contentType)

			assert.Equal(t, testCase.statusCode, resp.StatusCode())
		})
	}
}

func testRequest(t *testing.T, testServer *httptest.Server, method, path, contentType string) *resty.Response {
	t.Helper()

	client := resty.New()
	client.SetContentLength(true)

	request := client.R().SetHeader("Content-Type", contentType).SetBody("test")
	request.SetHeader("Content-Type", contentType)
	resp, err := request.Execute(method, testServer.URL+path)
	require.NoError(t, err)

	return resp
}
