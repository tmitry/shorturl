package handlers

import (
	"net/http"
	"strings"
)

const (
	MessageIncorrectURL  = "incorrect URL"
	MessageIncorrectUID  = "incorrect UID"
	MessageURLNotFound   = "URL not found"
	MessageIncorrectJSON = "incorrect JSON"

	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
)

func GetContentType(r *http.Response) string {
	s := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if i := strings.Index(s, ";"); i > -1 {
		s = s[0:i]
	}

	return s
}
