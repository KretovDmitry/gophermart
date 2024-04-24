package header

import (
	"net/http"
	"strings"
)

// IsTextPlainContentType returns true if the content type of the
// request is text/plain.
func IsTextPlainContentType(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.Index(contentType, ";"); i > -1 {
		contentType = contentType[0:i]
	}
	return contentType == "text/plain"
}

// IsApplicationJSONContentType returns true if the content type of the
// request is application/json.
func IsApplicationJSONContentType(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return contentType == "application/json"
}
