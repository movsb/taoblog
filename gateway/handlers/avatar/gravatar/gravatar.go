package gravatar

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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

type _Result struct {
	rsp *http.Response
	err error
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

	ch := make(chan _Result)

	// 并发请求所有源，错误的返回 nil
	for _, host := range gravatarHosts {
		go func(host string) {
			rsp, err := get(ctx, host, fmt.Sprintf(`%x`, sum))
			ch <- _Result{rsp, err}
		}(host)
	}

	var rsp *http.Response

	var errs []error

	// 只取第一个有效的
	for range gravatarHosts {
		ret := <-ch
		if ret.rsp == nil {
			errs = append(errs, ret.err)
			continue
		}

		if rsp == nil {
			rsp = ret.rsp
		} else {
			ret.rsp.Body.Close()
		}
	}

	if rsp == nil {
		return nil, fmt.Errorf(`no available gravatar: %v`, errors.Join(errs...))
	}

	return rsp, nil
}

func get(ctx context.Context, endpoint string, hash string) (*http.Response, error) {
	endpoint = strings.TrimSuffix(endpoint, `/`)
	u := fmt.Sprintf(`%s/%s?d=mm&s=100`, endpoint, hash)
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
		return nil, fmt.Errorf(`statusCode != 200: %s, %s`, u, resp.Status)
	}
}
