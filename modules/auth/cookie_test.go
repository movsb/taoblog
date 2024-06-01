package auth

import "testing"

func TestCookieGen(t *testing.T) {
	t.Log(cookieValue(`taoblog-ios-client/1.0`, `a`, `b`))
}
