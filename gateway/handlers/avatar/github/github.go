package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
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

func Get(ctx context.Context, email string) (_ *http.Response, outErr error) {
	defer utils.CatchAsError(&outErr)

	u := fmt.Sprintf(`https://api.github.com/search/users?q=%s`, email)
	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodGet, u, nil))
	resp := utils.Must1(http.DefaultClient.Do(req))
	if resp.StatusCode != 200 {
		return nil, errors.New(`statusCode != 200`)
	}
	defer resp.Body.Close()
	r := _UserSearchResult{}
	utils.Must(json.NewDecoder(resp.Body).Decode(&r))
	if len(r.Items) == 0 {
		return nil, errors.New(`no such github user`)
	}

	avatarURL := r.Items[0].AvatarURL

	req = utils.Must1(http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil))
	resp = utils.Must1(http.DefaultClient.Do(req))

	switch resp.StatusCode {
	case 200:
		return resp, nil
	default:
		return nil, errors.New(`statusCode != 200`)
	}
}
