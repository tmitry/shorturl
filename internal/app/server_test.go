package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app"
	"github.com/tmitry/shorturl/internal/app/handlers"
)

func TestRouter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
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
			name:        "incorrect POST path",
			method:      resty.MethodPost,
			path:        "/ds",
			statusCode:  http.StatusNotFound,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "incorrect Content-Type for POST /",
			method:      resty.MethodPost,
			path:        "/",
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "correct POST /",
			method:      resty.MethodPost,
			path:        "/",
			body:        "https://example.com",
			statusCode:  http.StatusCreated,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "not allowed method GET for /api/shorten",
			method:      resty.MethodPut,
			path:        "/api/shorten",
			statusCode:  http.StatusMethodNotAllowed,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "incorrect Content-Type for POST /api/shorten",
			method:      resty.MethodPost,
			path:        "/api/shorten",
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: handlers.ContentTypeText,
		},
		{
			name:        "correct POST /api/shorten",
			method:      resty.MethodPost,
			path:        "/api/shorten",
			body:        `{"url":"https://example.com/"}`,
			statusCode:  http.StatusCreated,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "incorrect Content-Encoding for POST /",
			method:      resty.MethodPost,
			path:        "/",
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: handlers.ContentTypeJSON,
		},
		{
			name:        "correct GET /api/user/urls",
			method:      resty.MethodGet,
			path:        "/api/user/urls",
			statusCode:  http.StatusNoContent,
			contentType: handlers.ContentTypeText,
		},
	}

	router := app.NewRouter()
	testServer := httptest.NewServer(router)

	t.Cleanup(func() {
		testServer.Close()
	})

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resp := testRequest(t, testServer, testCase.method, testCase.path, testCase.body, testCase.contentType, "")

			assert.Equal(t, testCase.statusCode, resp.StatusCode())
		})
	}
}

func TestAcceptEncoding(t *testing.T) {
	t.Parallel()

	type want struct {
		contentEncoding string
	}

	tests := []struct {
		name           string
		acceptEncoding string
		want           want
	}{
		{
			name:           "empty encoding",
			acceptEncoding: "",
			want: want{
				contentEncoding: "",
			},
		},
		{
			name:           "gzip encoding",
			acceptEncoding: "gzip",
			want: want{
				contentEncoding: "gzip",
			},
		},
		{
			name:           "deflate encoding",
			acceptEncoding: "deflate",
			want: want{
				contentEncoding: "deflate",
			},
		},
	}

	method := resty.MethodPost
	path := "/"
	body := "https://example.com/"
	contentType := handlers.ContentTypeText

	router := app.NewRouter()
	testServer := httptest.NewServer(router)

	t.Cleanup(func() {
		testServer.Close()
	})

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			resp := testRequest(t, testServer, method, path, body, contentType, testCase.acceptEncoding)

			assert.Equal(t, testCase.want.contentEncoding, resp.Header().Get("Content-Encoding"))
		})
	}
}

func testRequest(
	t *testing.T,
	testServer *httptest.Server,
	method, path, body, contentType, acceptEncoding string,
) *resty.Response {
	t.Helper()

	client := resty.New()
	client.SetContentLength(true)

	request := client.R()
	request.SetHeader("Content-Type", contentType)

	if acceptEncoding != "" {
		request.SetHeader("Accept-Encoding", acceptEncoding)
	}

	if body != "" {
		request.SetBody(body)
	}

	resp, err := request.Execute(method, testServer.URL+path)
	require.NoError(t, err)

	return resp
}
