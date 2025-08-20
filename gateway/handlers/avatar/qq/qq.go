package qq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// [QQ 居然暴露了通过 QQ 号直接获取头像的接口 - 陪她去流浪](https://blog.twofei.com/1809/)
// https://q.qlogo.cn/headimg_dl?dst_uin=191035066&spec=4
func Get(ctx context.Context, email string) (*http.Response, error) {
	before, after, found := strings.Cut(email, `@`)
	if !found || strings.ToLower(after) != `qq.com` {
		return nil, fmt.Errorf(`not qq email`)
	}
	uid, err := strconv.Atoi(before)
	if err != nil {
		return nil, fmt.Errorf(`not qq uid`)
	}
	url := fmt.Sprintf(`https://q.qlogo.cn/headimg_dl?dst_uin=%d&spec=4`, uid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil || rsp.StatusCode != 200 {
		return nil, errors.Join(err, fmt.Errorf(`status=%v`, rsp.Status))
	}
	return rsp, nil
}
