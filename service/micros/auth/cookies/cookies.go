package cookies

import (
	"crypto/sha1"
	"fmt"
	"log"
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
func CookieValue(userAgent, ip string, userID int, password string) string {
	return cookieValue(userAgent, ip, userID, password, time.Now())
}

// TODO 临时禁用IP参数校验。
// 获取真正的外部IP太麻烦、太不准确了。
// 有可能来自nginx、有可能来自直接访问、有可能来自内网访问，有可能来自docker proxy等等等。
func cookieValue(userAgent string, ip string, userID int, password string, t time.Time) string {
	_ = ip
	data := fmt.Sprintf(`%s,%s,%s,%d`, userAgent, "", password, t.Unix())
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

// 返回是否有效，是否应该刷新。
func ValidateCookieValue(value string, userAgent string, ip string, getUser func(userID int) (password string)) (bool, bool) {
	if userAgent == `` {
		return false, false
	}
	userID, _, tm := parseCookieValue(value)
	password := getUser(userID)

	expect := cookieValue(userAgent, ip, userID, password, tm)
	refresh := time.Since(tm) > maxAge/2
	return value == expect, refresh
}

var maxAge = time.Hour * 24

// 设置 Cookies 的最大有效期。
func SetMaxAge(d time.Duration) {
	maxAge = d
}

func isHTTPS(r *http.Request) bool {
	u, _ := url.Parse(r.Header.Get(`Origin`))
	return u != nil && strings.ToLower(u.Scheme) == `https`
}

func MakeCookie(w http.ResponseWriter, r *http.Request, userID int, password string, nickname string) {
	agent := r.Header.Get("User-Agent")
	ip := ParseRemoteAddrFromRequest(r)
	cookie := CookieValue(agent, ip.String(), userID, password)
	if (ip.IsLoopback() || ip.IsPrivate()) && !version.DevMode() {
		log.Println(`生成登录凭证：`, nickname, userID, ip)
		log.Printf("警告：登录 IP %s 是内网地址，部署可能有误。", ip)
	}
	secure := !version.DevMode() && isHTTPS(r)
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

func RemoveCookie(w http.ResponseWriter, r *http.Request) {
	secure := !version.DevMode() && isHTTPS(r)
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
