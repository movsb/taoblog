package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/modules/utils"
)

// 监控证书过期的剩余时间。
func (s *Service) monitorCert(chanify *notify.Chanify) {
	home := s.cfg.Site.Home
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
			if chanify != nil {
				chanify.Send(`错误`, err.Error(), true)
			}
			return
		}
		defer conn.Close()
		cert := conn.ConnectionState().PeerCertificates[0]
		left := time.Until(cert.NotAfter)
		if left <= 0 {
			log.Println(`已过期`)
			if chanify != nil {
				chanify.Send(`证书`, `已经过期。`, true)
			}
			return
		}
		daysLeft := int(left.Hours() / 24)
		s.certDaysLeft.Store(int32(daysLeft))
		if daysLeft >= 15 {
			return
		}
		log.Println(`剩余天数：`, daysLeft)
		if chanify != nil {
			chanify.Send(`证书`, fmt.Sprintf(`剩余天数：%v`, daysLeft), true)
		}
	}
	check()
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		defer ticker.Stop()
		for range ticker.C {
			check()
		}
	}()
}

// 监控域名过期的剩余时间。
func (s *Service) monitorDomain(chanify *notify.Chanify) {
	home := s.cfg.Site.Home
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

	check := func() {
		// curl --request GET \
		// --url 'https://api.apilayer.com/whois/query?domain=apilayer.com' \
		// --header 'apikey: YOUR API KEY HERE'
		u, err := url.Parse(`https://api.apilayer.com/whois/query?domain=`)
		if err != nil {
			log.Println(err)
			return
		}
		q := u.Query()
		q.Set(`domain`, domainSuffix)
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, u.String(), nil)
		if err != nil {
			log.Println(err)
			return
		}

		// 可以线上更改，所以总是重新取值。
		key := s.cfg.Others.Whois.ApiLayer.Key
		if key == "" {
			return
		}

		req.Header.Add(`apikey`, key)
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer rsp.Body.Close()
		if rsp.StatusCode != 200 {
			log.Println(`Status != 200`, rsp.StatusCode)
			return
		}
		var result struct {
			Result struct {
				DomainName     string `json:"domain_name"`
				ExpirationDate string `json:"expiration_date"`
			} `json:"result"`
		}
		if err := json.NewDecoder(rsp.Body).Decode(&result); err != nil {
			log.Println(err)
			return
		}
		// TODO 不知道时区。
		t, err := time.Parse(time.DateTime, result.Result.ExpirationDate)
		if err != nil {
			log.Println(err)
			return
		}

		daysLeft := time.Until(t) / time.Hour / 24
		s.domainExpirationDaysLeft.Store(int32(daysLeft))
		log.Println(`剩余天数：`, daysLeft)
		if chanify != nil && daysLeft < 15 {
			chanify.Send(`域名`, fmt.Sprintf(`剩余天数：%v`, daysLeft), true)
		}
	}
	// ApiLayer 限制是一个月 3000 次，这样可以做到
	// 即便不断重启，也会不超过限制。
	time.Sleep(time.Minute * 15)
	check()
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		defer ticker.Stop()
		for range ticker.C {
			check()
		}
	}()
}
