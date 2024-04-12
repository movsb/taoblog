package avatar

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	gGrAvatarHost = "https://www.gravatar.com/avatar"
)

func gravatar(_ context.Context, email string, p *Params) (*http.Response, error) {
	sum := sha256.Sum256([]byte(strings.ToLower(email)))
	u := fmt.Sprintf(`%s/%x?d=mm&s=100`, gGrAvatarHost, sum)
	log.Println(`请求头像：`, u)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range p.Headers {
		req.Header[k] = v
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200, 304:
		return resp, nil
	default:
		return nil, errors.New(`statusCode != 200`)
	}
}
