package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{
			name:       "not allowed method PUT",
			method:     resty.MethodPut,
			path:       "/",
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "not allowed method PATCH",
			method:     resty.MethodPatch,
			path:       "/",
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "empty GET path",
			method:     resty.MethodGet,
			path:       "/",
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "short GET path",
			method:     resty.MethodGet,
			path:       "/dsa",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "incorrect characters in GET path",
			method:     resty.MethodGet,
			path:       "/d$!sa",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "correct GET path",
			method:     resty.MethodGet,
			path:       "/J9KQP",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "incorrect POST path",
			method:     resty.MethodPost,
			path:       "/ds",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "incorrect POST path",
			method:     resty.MethodPost,
			path:       "/",
			statusCode: http.StatusBadRequest,
		},
	}

	router := NewRouter()
	testServer := httptest.NewServer(router)

	defer testServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, testServer, tt.method, tt.path)

			assert.Equal(t, tt.statusCode, resp.StatusCode())
		})
	}
}

func testRequest(t *testing.T, testServer *httptest.Server, method, path string) *resty.Response {
	t.Helper()

	client := resty.New()
	resp, err := client.R().Execute(method, testServer.URL+path)
	require.NoError(t, err)

	return resp
}
