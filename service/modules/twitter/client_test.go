package twitter

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
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

func TestUploadChunkedUsesV2Endpoints(t *testing.T) {
	step := 0
	httpClient := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		step++
		response := `{"data":{"id":"456"}}`
		switch step {
		case 1:
			if r.Method != http.MethodPost || r.URL.Path != "/2/media/upload/initialize" || r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected initialize request: %s %s %s", r.Method, r.URL.Path, r.Header.Get("Content-Type"))
			}
			var body struct {
				MediaType     string `json:"media_type"`
				MediaCategory string `json:"media_category"`
				TotalBytes    int    `json:"total_bytes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.MediaType != "video/mp4" || body.MediaCategory != "tweet_video" || body.TotalBytes != 3 {
				t.Fatalf("unexpected initialize body: %+v", body)
			}
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/2/media/upload/456/append" {
				t.Fatalf("unexpected append request: %s %s", r.Method, r.URL.Path)
			}
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || mediaType != "multipart/form-data" {
				t.Fatalf("unexpected append content type: %q: %v", mediaType, err)
			}
			reader := multipart.NewReader(r.Body, params["boundary"])
			form, err := reader.ReadForm(1 << 20)
			if err != nil {
				t.Fatal(err)
			}
			if got := form.Value["segment_index"]; !reflect.DeepEqual(got, []string{strconv.Itoa(0)}) {
				t.Fatalf("unexpected segment index: %v", got)
			}
			response = ""
		case 3:
			if r.Method != http.MethodPost || r.URL.Path != "/2/media/upload/456/finalize" {
				t.Fatalf("unexpected finalize request: %s %s", r.Method, r.URL.Path)
			}
		default:
			t.Fatalf("unexpected request number %d", step)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(response)),
			Header:     make(http.Header),
		}, nil
	})}

	id, err := NewClientWithHTTP(httpClient, "https://api.twitter.test").UploadChunked(
		context.Background(), "video/mp4", "tweet_video", []byte("abc"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if id != "456" || step != 3 {
		t.Fatalf("id=%q step=%d", id, step)
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
