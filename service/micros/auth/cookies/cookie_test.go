package cookies

import (
	"testing"
)

func TestCookieGen(t *testing.T) {
	ua := `taoblog-ios-client/1.0`
	v := CookieValue(ua, `127.0.0.1`, 2, `b`)
	valid, refresh := ValidateCookieValue(v, ua, `127.0.0.1`, func(userID int) (password string) {
		if userID == 2 {
			return `b`
		}
		return ``
	})
	t.Log(`cookie:`, v)
	if !valid {
		t.Fatal(valid, refresh)
	}
}
