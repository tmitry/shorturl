package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	MessageIncorrectURL    = "incorrect URL"
	MessageIncorrectUID    = "incorrect UID"
	MessageURLNotFound     = "URL not found"
	MessageIncorrectJSON   = "incorrect JSON"
	MessageIncorrectUserID = "incorrect user ID"
	MessageURLWasDeleted   = "URL was deleted"

	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
	ContentTypeGZIP = "application/x-gzip"

	ContentEncodingGZIP = "gzip"
)

func GetContentType(r *http.Response) string {
	s := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if i := strings.Index(s, ";"); i > -1 {
		s = s[0:i]
	}

	return s
}

func getRequestReader(request *http.Request) (io.ReadCloser, error) {
	if request.Header.Get("Content-Encoding") == ContentEncodingGZIP {
		gzipReader, err := gzip.NewReader(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}

		return gzipReader, nil
	}

	return request.Body, nil
}
