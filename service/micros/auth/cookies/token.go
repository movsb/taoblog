package cookies

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
)

const TokenName = `token`

func TokenValue(userID int, password string) string {
	return fmt.Sprintf(`%d:%s`, userID, shasum(password))
}

func AuthorizationValue(userID int, password string) string {
	return TokenName + ` ` + TokenValue(userID, password)
}

// a: token id:sum
// returns: [<id>,<id:sum>]
func ParseAuthorization(a string) (int, string, bool) {
	splits := strings.Fields(a)
	if len(splits) != 2 {
		return 0, "", false
	}
	if splits[0] != TokenName {
		return 0, "", false
	}
	token := splits[1]
	splits = strings.Split(token, `:`)
	if len(splits) != 2 {
		return 0, "", false
	}
	id, err := strconv.Atoi(splits[0])
	if err != nil {
		log.Println(err)
		return 0, "", false
	}

	return id, token, true
}

// NOTE x-forwarded-for 可能是伪造的
// 但是我现在只有一层nginx，且在nginx那边配置了 proxy_set_headers 来覆盖掉用户伪造的 x-forwarded-for，所以这里就不再验证了。
func ParseRemoteAddrFromRequest(r *http.Request) netip.Addr {
	var f string
	if fs := r.Header.Values(`x-forwarded-for`); len(fs) > 0 {
		f = fs[0]
	}
	if f == "" {
		f, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ParseRemoteAddr(f)
}

func ParseRemoteAddr(f string) netip.Addr {
	if f == "" {
		panic(`缺少 X-Forwarded-For / RemoteAddr / Peer 字段。`)
	}
	if p := strings.IndexByte(f, ','); p != -1 {
		f = f[:p]
	}
	return netip.MustParseAddr(f)
}
