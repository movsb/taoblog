package service

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/calendar"
)

type CalendarData struct {
	UserID int `json:"user_id"`
}

// 请求来自日历软件， r 是不带 cookies 鉴权的。
// 需要自行从 query 鉴权。
func (s *Service) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	encoded := r.URL.Query().Get(`data`)
	encrypted, _ := base64.RawURLEncoding.DecodeString(encoded)
	decrypted, err := s.aesGCM.Decrypt([]byte(encrypted))
	if err != nil {
		log.Println(`cannot decrypt:`, encrypted)
		http.Error(w, `bad data`, 400)
		return
	}

	var cd CalendarData

	if err := json.Unmarshal(decrypted, &cd); err != nil {
		log.Println(`error unmarshaling:`, err, decrypted)
		http.Error(w, err.Error(), 500)
		return
	}

	calendar.AddHeaders(w)

	events := s.calendar.Filter(func(e *calendar.Event) bool {
		// 管理员获取全部，其它用户获取自己。
		return cd.UserID == auth.AdminID || cd.UserID == e.UserID
	})

	events = events.Unique(func(e *calendar.Event) string {
		uuid := e.Tags[`uuid`].(string)
		repeat, _ := e.Tags[`repeat`].(string)
		return uuid + repeat
	})

	info, _ := s.GetInfo(r.Context(), &proto.GetInfoRequest{})
	events.Marshal(info.Name, w)
}
