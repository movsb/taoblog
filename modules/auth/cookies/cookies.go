package cookies

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/version"
)

const (
	CookieNameLogin    = `taoblog.login`
	CookieNameUserID   = `taoblog.user_id`
	CookieNameNickname = `taoblog.nickname`
)

func shasum(in string) string {
	h := sha1.Sum([]byte(in))
	return fmt.Sprintf("%x", h)
}

// 生成一个与当前时间相关的 Cookie 值。
func CookieValue(userAgent string, userID int, password string) string {
	return cookieValue(userAgent, userID, password, time.Now())
}

func cookieValue(userAgent string, userID int, password string, t time.Time) string {
	data := fmt.Sprintf(`%s,%s,%d`, userAgent, password, t.Unix())
	sum := shasum(data)
	return fmt.Sprintf(`%d:%s:%d`, userID, sum, t.Unix())
}

func parseCookieValue(value string) (userID int, sum string, tm time.Time) {
	parts := strings.Split(value, `:`)
	if len(parts) == 3 {
		userID, _ = strconv.Atoi(parts[0])
		sum = parts[1]
		t, _ := strconv.Atoi(parts[2])
		tm = time.Unix(int64(t), 0)
	}
	return
}

func ValidateCookieValue(value string, userAgent string, getUser func(userID int) (password string)) bool {
	if userAgent == `` {
		return false
	}
	userID, _, tm := parseCookieValue(value)
	password := getUser(userID)
	expect := cookieValue(userAgent, userID, password, tm)
	return value == expect
}

const maxAge = time.Hour * 24 * 7

func MakeCookie(w http.ResponseWriter, r *http.Request, userID int, password string, nickname string) {
	agent := r.Header.Get("User-Agent")
	cookie := CookieValue(agent, userID, password)
	secure := !version.DevMode()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    cookie,
		MaxAge:   int(maxAge.Seconds()),
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	// 只用于前端展示使用，不能用作凭证。
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameUserID,
		Value:    fmt.Sprint(userID),
		MaxAge:   int(maxAge.Seconds()),
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameNickname,
		Value:    url.PathEscape(nickname),
		MaxAge:   int(maxAge.Seconds()),
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func RemoveCookie(w http.ResponseWriter) {
	secure := !version.DevMode()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameLogin,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameUserID,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameNickname,
		Value:    ``,
		MaxAge:   -1,
		Path:     `/`,
		Domain:   ``,
		Secure:   secure,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func RefreshCookies(w http.ResponseWriter, r *http.Request) {

}
