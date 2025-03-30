package server

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

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// 监控证书过期的剩余时间。
func monitorCert(ctx context.Context, home string, notifier proto.NotifyServer, update func(days int)) {
	u, err := url.Parse(home)
	if err != nil {
		panic(err)
	}
	if u.Scheme != `https` {
		return
	}
	port := utils.IIF(u.Port() == "", "443", u.Port())
	addr := net.JoinHostPort(u.Hostname(), port)
	check := func() {
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
				auth.SystemForLocal(ctx),
				&proto.SendInstantRequest{
					Subject: `证书`,
					Body:    `已经过期`,
				},
			)
		}
		daysLeft := int(left.Hours() / 24)
		// s.certDaysLeft.Store(int32(daysLeft))
		// s.exporter.certDaysLeft.Set(float64(daysLeft))
		update(daysLeft)
		if daysLeft >= 15 {
			return
		}
		log.Println(`证书剩余天数：`, daysLeft)
		notifier.SendInstant(
			auth.SystemForLocal(ctx),
			&proto.SendInstantRequest{
				Subject: `证书`,
				Body:    fmt.Sprintf(`剩余天数：%v`, daysLeft),
			},
		)
	}
	check()

	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()
	for range ticker.C {
		check()
	}
}

// 监控域名过期的剩余时间。
func monitorDomain(ctx context.Context, home string, notifier proto.NotifyServer, apiKey string, update func(days int)) {
	u, err := url.Parse(home)
	if err != nil {
		panic(err)
	}
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
		log.Println(`没有已知的域名后缀。`)
		return
	}

	domainSuffix := strings.Join(suffix, ".")

	check := func() error {
		// curl --request GET \
		// --url 'https://api.apilayer.com/whois/query?domain=apilayer.com' \
		// --header 'apikey: YOUR API KEY HERE'
		u, err := url.Parse(`https://api.apilayer.com/whois/query?domain=`)
		if err != nil {
			log.Println(err)
			return err
		}
		q := u.Query()
		q.Set(`domain`, domainSuffix)
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, u.String(), nil)
		if err != nil {
			log.Println(err)
			return err
		}

		// 可以线上更改，所以总是重新取值。
		if apiKey == "" {
			return errors.New(`no key specified`)
		}

		req.Header.Add(`apikey`, apiKey)
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			return err
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != 200 {
			log.Println(`Status != 200`, rsp.StatusCode)
			return err
		}
		var result struct {
			Result struct {
				DomainName     string `json:"domain_name"`
				ExpirationDate string `json:"expiration_date"`
			} `json:"result"`
		}
		if err := json.NewDecoder(rsp.Body).Decode(&result); err != nil {
			log.Println(err)
			return err
		}
		// TODO 不知道时区。
		t, err := time.Parse(time.DateTime, result.Result.ExpirationDate)
		if err != nil {
			log.Println(err)
			return err
		}

		daysLeft := int(time.Until(t) / time.Hour / 24)
		update(daysLeft)
		log.Println(`域名剩余天数：`, daysLeft)
		if daysLeft < 15 {
			notifier.SendInstant(
				auth.SystemForLocal(ctx),
				&proto.SendInstantRequest{
					Subject: `域名`,
					Body:    fmt.Sprintf(`剩余天数：%v`, daysLeft),
				},
			)
		}

		return nil
	}
	// ApiLayer 限制是一个月 3000 次，这样可以做到
	// 即便不断重启，也会不超过限制。
	time.Sleep(time.Minute * 15)
	check()
	time.Sleep(time.Minute * 15)

	for {
		if check() == nil {
			time.Sleep(time.Hour * 24)
		} else {
			time.Sleep(time.Minute * 15)
		}
	}
}
