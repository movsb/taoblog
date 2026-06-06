package expiration

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/whois"
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
func MonitorDomain(ctx context.Context, getHome func() string, notifier proto.NotifyServer, initialDelay bool, update func(days int)) {
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
		suffix := getDomainSuffix()
		if suffix == `` {
			return errors.New(`无法获取域名后缀`)
		}

		t, err := whois.QueryDomainExpiration(suffix)
		if err != nil {
			return fmt.Errorf(`查询域名过期时间错误：%v`, err)
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

	// 即便由于代码问题程序不断重启，也不会超过请求限制。
	if initialDelay {
		time.Sleep(time.Minute)
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
