package avatar

import (
	"context"
	"net/http"
)

// Params ...
type Params struct {
	Headers http.Header
}

// Get ...
func Get(email string, p *Params) (*http.Response, error) {
	var err error
	var resp *http.Response
	resp, err = github(context.TODO(), email, p)
	if err != nil {
		resp, err = gravatar(context.TODO(), email, p)
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}
