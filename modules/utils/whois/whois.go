package whois

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// 写完手动查询发现Copilot提示了一个免费的？
// https://api.whois.vu/?q=twofei.com

func QueryDomainExpiration(domain string) (time.Time, error) {
	server := `whois.iana.org`

	for {
		kv, err := query(server, domain)
		if err != nil {
			return time.Time{}, err
		}
		if refer, ok := kv[`refer`]; ok {
			server = refer
			continue
		}
		if exp, ok := kv[`Registry Expiry Date`]; ok {
			for _, layout := range []string{
				`2006-01-02 15:04:05-07:00`,
				`2006-01-02 15:04:05`,
				`2006-01-02T15:04:05-07:00`,
				`2006-01-02T15:04:05`,
				time.RFC3339,
			} {
				t, err := time.Parse(layout, exp)
				if err == nil {
					return t, nil
				}
			}
			return time.Time{}, fmt.Errorf(`无法解析日期：%v`, exp)
		}
		return time.Time{}, fmt.Errorf(`无法找到域名过期日期`)
	}
}

func query(server string, q string) (map[string]string, error) {
	conn, err := net.Dial(`tcp4`, server+`:43`)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(q + "\r\n")); err != nil {
		return nil, err
	}

	scn := bufio.NewScanner(conn)
	mss := map[string]string{}
	for scn.Scan() {
		line := scn.Text()
		line = strings.TrimSpace(line)
		if before, after, ok := strings.Cut(line, ":"); ok {
			key := strings.TrimSpace(before)
			val := strings.TrimSpace(after)
			mss[key] = val
		}
	}

	return mss, scn.Err()
}
