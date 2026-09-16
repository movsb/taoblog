package twitter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestCreatePost(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/2/tweets" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Text  string `json:"text"`
			Media struct {
				IDs []string `json:"media_ids"`
			} `json:"media"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Text != "hello" || !reflect.DeepEqual(body.Media.IDs, []string{"10", "11"}) {
			t.Fatalf("unexpected body: %+v", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"data":{"id":"123"}}`)),
			Header:     make(http.Header),
		}, nil
	})}

	client := NewClientWithHTTP(httpClient, "https://api.x.test")
	id, err := client.CreatePost(context.Background(), "hello", []string{"10", "11"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "123" {
		t.Fatalf("got id %q", id)
	}
}

func TestAPIErrorRetryable(t *testing.T) {
	for _, tc := range []struct {
		status    int
		retryable bool
	}{
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
	} {
		got := (&APIError{StatusCode: tc.status}).Retryable()
		if got != tc.retryable {
			t.Fatalf("status %d: got %v", tc.status, got)
		}
	}
}
