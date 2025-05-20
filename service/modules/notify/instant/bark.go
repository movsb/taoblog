package instant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/movsb/taoblog/modules/utils"
)

type Level string

const (
	Critical      Level = `critical`
	Active        Level = `active`
	TimeSensitive Level = `timeSensitive`
	Passive       Level = `passive`
)

type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`

	Level Level  `json:"level"`
	Group string `json:"group"`
	URL   string `json:"url"`
}

type _Message struct {
	Message `json:",inline"`

	DeviceKey string
}

// https://blog.twofei.com/1683/
type _BarkMessage struct {
	Message `json:",inline"`

	Action string `json:"action"`
	Badge  int    `json:"badge"`
}

const endpoint = `https://api.day.app`

type _Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
}

func SendBarkMessage(ctx context.Context, deviceKey string, m Message) error {
	u := utils.Must1(url.Parse(endpoint)).JoinPath(deviceKey)
	body := utils.Must1(json.Marshal(_BarkMessage{
		Message: m,
		Action:  `none`,
		Badge:   1, // 随意设置的
	}))
	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body)))
	req.Header.Set(`Content-Type`, `application/json`)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(`SendBarkMessage: %s %+v: %w`, deviceKey, m, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return fmt.Errorf(`SendBarkMessage: status=%q`, rsp.Status)
	}
	var r _Response
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return fmt.Errorf(`SendBarkMessage: bad response: %w`, err)
	}
	if r.Code != 200 {
		return fmt.Errorf(`SendBarkMessage: bad code: %d`, r.Code)
	}
	return nil
}
