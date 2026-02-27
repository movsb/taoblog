package expiration

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

// 监控证书过期的剩余时间。
//
// 同步函数，除非 ctx 结束，否则不会返回。
func MonitorCert(ctx context.Context, getHome func() string, notifier proto.NotifyServer, update func(days int)) {
	check := func() {
		u := utils.Must1(url.Parse(getHome()))
		if u.Scheme != `https` {
			return
		}
		port := utils.IIF(u.Port() == "", "443", u.Port())
		addr := net.JoinHostPort(u.Hostname(), port)
		conn, err := tls.Dial(`tcp`, addr, &tls.Config{})
		if err != nil {
			log.Println(err)
			// notifier.Notify(`错误`, err.Error())
			return
		}
		defer conn.Close()
		cert := conn.ConnectionState().PeerCertificates[0]
		left := time.Until(cert.NotAfter)
		if left <= 0 {
			log.Println(`已过期`)
			notifier.SendInstant(
				user.SystemForLocal(ctx),
				&proto.SendInstantRequest{
					Title: `证书`,
					Body:  `已经过期`,
					Group: `系统状态`,
					Level: proto.SendInstantRequest_Passive,
				},
			)
		}
		daysLeft := int(left.Hours() / 24)
		update(daysLeft)
		if daysLeft >= 15 {
			return
		}
		log.Println(`证书剩余天数：`, daysLeft)
		notifier.SendInstant(
			user.SystemForLocal(ctx),
			&proto.SendInstantRequest{
				Title: `证书`,
				Body:  fmt.Sprintf(`剩余天数：%v`, daysLeft),
				Group: `系统状态`,
				Level: proto.SendInstantRequest_Passive,
			},
		)
	}

	check()

	utils.AtMiddleNight(ctx, check)
}

// 监控域名过期的剩余时间。
//
// 同步函数，除非 ctx 结束，否则不会返回。
//
// notifier 可以为空。
func MonitorDomain(ctx context.Context, getHome func() string, notifier proto.NotifyServer, apiKey string, initialDelay bool, update func(days int)) {
	getDomainSuffix := func() string {
		u := utils.Must1(url.Parse(getHome()))
		hostname := strings.ToLower(u.Hostname())
		fields := strings.Split(hostname, `.`)
		suffix := []string{}
		switch fields[len(fields)-1] {
		case `com`:
			if len(fields) >= 2 {
				suffix = fields[len(fields)-2:]
			}
		}
		if len(suffix) <= 0 {
			log.Println(`没有已知的域名后缀：`, getHome())
			return ``
		}
		return strings.Join(suffix, ".")
	}

	check := func() error {
		// curl --request GET \
		// --url 'https://api.apilayer.com/whois/query?domain=apilayer.com' \
		// --header 'apikey: YOUR API KEY HERE'
		u, err := url.Parse(`https://api.apilayer.com/whois/query?domain=`)
		if err != nil {
			return err
		}
		q := u.Query()
		q.Set(`domain`, getDomainSuffix())
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}

		// 可以线上更改，所以总是重新取值。
		if apiKey == "" {
			return errors.New(`no key specified`)
		}

		req.Header.Add(`apikey`, apiKey)
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != 200 {
			return fmt.Errorf(`status != 200: %d`, rsp.StatusCode)
		}
		var result struct {
			Result struct {
				DomainName     string `json:"domain_name"`
				ExpirationDate string `json:"expiration_date"`
			} `json:"result"`
		}
		if err := json.NewDecoder(rsp.Body).Decode(&result); err != nil {
			return err
		}

		var (
			t        time.Time
			parseErr error
		)
		for _, layout := range []string{
			`2006-01-02 15:04:05-07:00`,
			`2006-01-02 15:04:05`,
			`2006-01-02T15:04:05-07:00`,
			`2006-01-02T15:04:05`,
		} {
			t2, err := time.Parse(layout, result.Result.ExpirationDate)
			if err == nil {
				t = t2
				parseErr = nil
				break
			}
			parseErr = errors.Join(parseErr, err)
		}
		if parseErr != nil {
			return parseErr
		}

		daysLeft := int(time.Until(t) / time.Hour / 24)
		update(daysLeft)
		if daysLeft < 15 {
			log.Println(`域名剩余天数：`, daysLeft)
			if notifier != nil {
				notifier.SendInstant(
					user.SystemForLocal(ctx),
					&proto.SendInstantRequest{
						Title: `域名`,
						Body:  fmt.Sprintf(`剩余天数：%v`, daysLeft),
						Group: `系统状态`,
						Level: proto.SendInstantRequest_Passive,
					},
				)
			}
		}

		return nil
	}

	// ApiLayer 限制是一个月 3000 次，这样可以做到
	// 即便由于代码问题程序不断重启，也不会超过请求限制。
	if initialDelay {
		time.Sleep(time.Minute * 15)
	}

	if err := check(); err != nil {
		log.Println(err)
	}

	utils.AtMiddleNight(ctx, func() {
		if err := check(); err != nil {
			log.Println(err)
		}
	})
}
