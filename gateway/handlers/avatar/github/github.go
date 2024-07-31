package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Items []struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"items"`
}

func Get(ctx context.Context, email string) (*http.Response, error) {
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
	if len(r.Items) == 0 {
		return nil, errors.New(`no such github user`)
	}

	avatarURL := r.Items[0].AvatarURL

	req, err := http.NewRequest(http.MethodGet, avatarURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		return resp, nil
	default:
		return nil, errors.New(`statusCode != 200`)
	}
}
