package gravatar

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// NOTE: 第一个是主站，其它的都是镜像站。
//
// 镜像站的 Last-Modified, ETag 理论上和主站一样，所以可以共用缓存。
//
// 一些来源：
//
// - https://zhuanlan.zhihu.com/p/377149911
// - https://github.com/eallion/cloudflare-gravatar
//
// 黑名单：
//
// - WeAvatar <https://blog.twofei.com/902/#comment-1593>
var gravatarHosts = []string{
	`https://www.gravatar.com/avatar`,
	`https://sdn.geekzu.org/avatar`,
	`https://cdn.v2ex.com/gravatar`,
	`https://gravatar.loli.net/avatar`,
	`https://cdn.sep.cc/avatar`,
	`https://cravatar.eallion.com/avatar`,
}

// 鉴于博客服务器可能在国外/可能在国内、主站可能被墙、镜像站不稳定的原因，
// 采用随机并发请求多个源的方式，外部做好缓存。
func Get(ctx context.Context, email string) (*http.Response, error) {
	if len(gravatarHosts) <= 0 {
		return nil, fmt.Errorf(`no gravatar host available`)
	}

	sum := sha256.Sum256([]byte(strings.ToLower(email)))

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	ch := make(chan *http.Response)

	// 并发请求所有源，错误的返回 nil
	for _, host := range gravatarHosts {
		go func(host string) {
			rsp, err := get(ctx, host, sum)
			if err != nil {
				log.Println(err)
				ch <- nil
			} else {
				ch <- rsp
			}
		}(host)
	}

	var rsp *http.Response

	// 只取第一个有效的
	for range gravatarHosts {
		rsp2 := <-ch
		if rsp2 == nil {
			continue
		}

		if rsp == nil {
			rsp = rsp2
		} else {
			rsp2.Body.Close()
		}
	}

	if rsp == nil {
		return nil, fmt.Errorf(`no available gravatar`)
	}

	return rsp, nil
}

func get(ctx context.Context, endpoint string, hash [32]byte) (*http.Response, error) {
	// TODO: 没有检测 endpoint 是否合法。
	u := fmt.Sprintf(`%s/%x?d=mm&s=100`, endpoint, hash)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		return resp, nil
	default:
		return nil, fmt.Errorf(`statusCode != 200: %s`, u)
	}
}
