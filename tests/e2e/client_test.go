package e2e_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestClientLogin(t *testing.T) {
	r := Serve(t.Context())

	// 获取授权链接。
	sc := utils.Must1(r.client.ClientLogin.ClientLogin(r.user1, &proto.ClientLoginRequest{}))
	rsp := utils.Must1(sc.Recv())
	authURL := rsp.GetOpen().GetAuthUrl()
	random := utils.Must1(url.Parse(authURL)).Query().Get(`random`)

	// 模拟授权按钮。
	adminURL := utils.Must1(url.Parse(r.server.JoinPath(`/admin/login/client`)))
	query := adminURL.Query()
	query.Set(`random`, random)
	adminURL.RawQuery = query.Encode()
	req := utils.Must1(http.NewRequest(http.MethodPost, adminURL.String(), nil))
	r.addAuth(req, r.user1ID)
	httpRsp := utils.Must1(http.DefaultClient.Do(req))
	if httpRsp.StatusCode != 200 {
		t.Fatal(`状态码不为 200`)
	}

	// 接收授权码。
	rsp = utils.Must1(sc.Recv())
	t.Log(rsp.GetSuccess().Token)
}
