package avatar

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultAvatar(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /v3/avatar/{id}", New(nil, nil).Handler())
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v3/avatar/0", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != defaultContentType || !bytes.Equal(response.Body.Bytes(), defaultAvatar) {
		t.Fatalf("default avatar response: status=%d type=%q bytes=%d", response.Code, response.Header().Get("Content-Type"), response.Body.Len())
	}
}
