package avatar

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
)

const (
	gGrAvatarHost = "https://www.gravatar.com/avatar"
)

func gravatar(ctx context.Context, email string, p *Params) (*http.Response, error) {
	u := fmt.Sprintf(`%s/%s?d=mm&s=48`, gGrAvatarHost, utils.Md5Str(email))
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
