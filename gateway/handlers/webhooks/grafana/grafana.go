package grafana

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/status"
)

func New(auther *auth.Auth, notify proto.NotifyClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.MaxBytesReader(w, r.Body, 1<<20)
		defer rc.Close()
		body := utils.DropLast1(io.ReadAll(rc))
		var m map[string]any
		json.Unmarshal(body, &m)
		var message string
		if x, ok := m[`message`]; ok {
			message, _ = x.(string)
		}
		ctx := auther.NewContextForRequestAsGateway(r)
		_, err := notify.SendInstant(ctx, &proto.SendInstantRequest{
			Subject: `监控告警`,
			// https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
			Body: message,
		})
		if err != nil {
			http.Error(w, err.Error(), runtime.HTTPStatusFromCode(status.Code(err)))
			return
		}
	})
}
