package cookies

import "testing"

func TestCookieGen(t *testing.T) {
	ua := `taoblog-ios-client/1.0`
	v := CookieValue(ua, 2, `b`)
	ok := ValidateCookieValue(v, ua, func(userID int) (password string) {
		if userID == 2 {
			return `b`
		}
		return ``
	})
	t.Log(`cookie:`, v)
	if !ok {
		t.Fatal(ok)
	}
}
