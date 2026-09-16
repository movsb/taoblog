package twitter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
)

type Config struct {
	Enabled           bool
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	APIBaseURL        string
}

func (c Config) Valid() bool {
	return c.Enabled && c.ConsumerKey != "" && c.ConsumerSecret != "" && c.AccessToken != "" && c.AccessTokenSecret != ""
}

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient(config Config) *Client {
	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessToken, config.AccessTokenSecret)
	return NewClientWithHTTP(oauthConfig.Client(context.Background(), token), config.APIBaseURL)
}

func NewClientWithHTTP(client *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.x.com"
	}
	return &Client{http: client, baseURL: strings.TrimRight(baseURL, "/")}
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Twitter API returned HTTP %d: %s", e.StatusCode, e.Body)
}

func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func (c *Client) do(req *http.Request, out any) error {
	rsp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(rsp.Body, 1<<20))
	if err != nil {
		return err
	}
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return &APIError{StatusCode: rsp.StatusCode, Body: strings.TrimSpace(string(body))}
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode Twitter API response: %w", err)
		}
	}
	return nil
}

func (c *Client) json(ctx context.Context, method, endpoint string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

type createPostResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *Client) CreatePost(ctx context.Context, text string, mediaIDs []string) (string, error) {
	body := map[string]any{"text": text}
	if len(mediaIDs) > 0 {
		body["media"] = map[string]any{"media_ids": mediaIDs}
	}
	var out createPostResponse
	if err := c.json(ctx, http.MethodPost, "/2/tweets", body, &out); err != nil {
		return "", err
	}
	if out.Data.ID == "" {
		return "", fmt.Errorf("Twitter API create post response has no id")
	}
	return out.Data.ID, nil
}

type mediaResponse struct {
	Data struct {
		ID             string `json:"id"`
		MediaIDString  string `json:"media_id_string"`
		ProcessingInfo *struct {
			State          string `json:"state"`
			CheckAfterSecs int    `json:"check_after_secs"`
			Error          *struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"processing_info"`
	} `json:"data"`
	MediaIDString string `json:"media_id_string"`
}

func (r mediaResponse) id() string {
	if r.Data.ID != "" {
		return r.Data.ID
	}
	if r.Data.MediaIDString != "" {
		return r.Data.MediaIDString
	}
	return r.MediaIDString
}

func (c *Client) UploadImage(ctx context.Context, contentType string, data []byte) (string, error) {
	body := map[string]any{
		"media":          base64.StdEncoding.EncodeToString(data),
		"media_category": "tweet_image",
		"media_type":     contentType,
	}
	var out mediaResponse
	if err := c.json(ctx, http.MethodPost, "/2/media/upload", body, &out); err != nil {
		return "", err
	}
	if out.id() == "" {
		return "", fmt.Errorf("Twitter API upload image response has no media id")
	}
	return out.id(), nil
}

func (c *Client) form(ctx context.Context, method, endpoint string, values url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *Client) appendChunk(ctx context.Context, mediaID string, index int, data []byte) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("command", "APPEND")
	_ = w.WriteField("media_id", mediaID)
	_ = w.WriteField("segment_index", strconv.Itoa(index))
	part, err := w.CreateFormFile("media", "chunk")
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/2/media/upload", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req, nil)
}

func (c *Client) UploadChunked(ctx context.Context, contentType, category string, data []byte) (string, error) {
	var init mediaResponse
	err := c.form(ctx, http.MethodPost, "/2/media/upload", url.Values{
		"command":        {"INIT"},
		"total_bytes":    {strconv.Itoa(len(data))},
		"media_type":     {contentType},
		"media_category": {category},
	}, &init)
	if err != nil {
		return "", err
	}
	mediaID := init.id()
	if mediaID == "" {
		return "", fmt.Errorf("Twitter API initialize upload response has no media id")
	}
	const chunkSize = 4 << 20
	for start, index := 0, 0; start < len(data); start, index = start+chunkSize, index+1 {
		end := min(start+chunkSize, len(data))
		if err := c.appendChunk(ctx, mediaID, index, data[start:end]); err != nil {
			return "", err
		}
	}
	var final mediaResponse
	if err := c.form(ctx, http.MethodPost, "/2/media/upload", url.Values{
		"command":  {"FINALIZE"},
		"media_id": {mediaID},
	}, &final); err != nil {
		return "", err
	}
	info := final.Data.ProcessingInfo
	for info != nil && info.State != "succeeded" {
		if info.State == "failed" {
			message := "media processing failed"
			if info.Error != nil && info.Error.Message != "" {
				message = info.Error.Message
			}
			return "", fmt.Errorf("Twitter API: %s", message)
		}
		wait := time.Duration(max(info.CheckAfterSecs, 1)) * time.Second
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(wait):
		}
		var status mediaResponse
		endpoint := "/2/media/upload?command=STATUS&media_id=" + url.QueryEscape(mediaID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
		if err != nil {
			return "", err
		}
		if err := c.do(req, &status); err != nil {
			return "", err
		}
		info = status.Data.ProcessingInfo
	}
	return mediaID, nil
}
