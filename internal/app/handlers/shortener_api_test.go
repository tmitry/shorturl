package handlers

import (
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

func TestShortenerAPIHandler_Shorten(t *testing.T) {
	t.Parallel()

	rep := repositories.NewMemoryRepository()
	handler := NewShortenerAPIHandler(rep)

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
				contentType: ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectJSON),
			},
		},
		{
			name: "bad JSON format",
			json: `{"ur_l":"https://example.com/"}`,
			want: want{
				contentType: ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			},
		},
		{
			name: "bad URL",
			json: `{"url":"bad_url"}`,
			want: want{
				contentType: ContentTypeText,
				statusCode:  http.StatusBadRequest,
				content:     fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), messageIncorrectURL),
			},
		},
		{
			name: "correct URL",
			json: `{"url":"https://example.com/"}`,
			want: want{
				contentType: ContentTypeJSON,
				statusCode:  http.StatusCreated,
				content:     "",
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPost, config.ServerCfg.Address, strings.NewReader(testCase.json))
			request.Header.Set("Content-Type", ContentTypeJSON)

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