package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	gGrAvatarHost = "https://www.gravatar.com/avatar/"
)

// GetAvatar gets avatar from GRAVATAR.
func GetAvatar(c *gin.Context) {
	query := c.Request.URL.RawQuery
	query, _ = url.QueryUnescape(query)
	ifModified := c.GetHeader("If-Modified-Since")
	ifNoneMatch := c.GetHeader("If-None-Match")

	u := gGrAvatarHost + query

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}

	if ifModified != "" {
		req.Header.Set("If-Modified-Since", ifModified)
	}
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	c.Status(resp.StatusCode)

	for name, value := range resp.Header {
		c.Header(name, value[0])
	}

	io.Copy(c.Writer, resp.Body)
}
