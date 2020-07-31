package avatar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// https://api.github.com/search/users?q=anhbk@qq.com
/*
{
  "total_count": 1,
  "incomplete_results": false,
  "items": [
    {
      "avatar_url": "https://avatars1.githubusercontent.com/u/5374525?v=4",
    }
  ]
}
*/

type _UserSearchResult struct {
	Iterms []struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"items"`
}

var (
	githubEmailCache = make(map[string]string)
	githubLock       sync.RWMutex
)

func github(ctx context.Context, email string, p *Params) (*http.Response, error) {
	avartarURL := ``

	githubLock.RLock()
	avartarURL = githubEmailCache[email]
	githubLock.RUnlock()

	if avartarURL == `` {
		u := fmt.Sprintf(`https://api.github.com/search/users?q=%s`, email)
		resp, err := http.Get(u)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New(`statusCode != 200`)
		}
		defer resp.Body.Close()
		r := _UserSearchResult{}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, err
		}
		if len(r.Iterms) == 0 {
			return nil, errors.New(`no such github user`)
		}

		avartarURL = r.Iterms[0].AvatarURL

		githubLock.Lock()
		githubEmailCache[email] = avartarURL
		githubLock.Unlock()
	}

	req, err := http.NewRequest(http.MethodGet, avartarURL, nil)
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
