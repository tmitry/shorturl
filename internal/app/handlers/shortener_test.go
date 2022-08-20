package handlers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

func TestShortenerHandler_ServeHTTP(t *testing.T) {
	rep := repositories.NewMemoryRepository()
	handler := http.Handler(NewShortenerHandler(rep))

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
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			},
		},
		{
			name: "correct URL",
			url:  "https://example.com/",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
				content:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, config.DefaultAddr, bytes.NewReader([]byte(tt.url)))

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)
			result := recorder.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			content, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if result.StatusCode != http.StatusCreated {
				assert.Equal(t, tt.want.content, strings.TrimSuffix(string(content), "\n"))
			}
		})
	}
}
