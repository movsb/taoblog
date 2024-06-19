package grafana

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/status"
)

func New(client proto.UtilsServer) http.Handler {
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
		_, err := client.InstantNotify(r.Context(), &proto.InstantNotifyRequest{
			Title: `监控告警`,
			// https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
			Message: message,
		})
		if err != nil {
			http.Error(w, err.Error(), runtime.HTTPStatusFromCode(status.Code(err)))
			return
		}
	})
}
