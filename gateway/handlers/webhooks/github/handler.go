package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type _GitHub struct {
	client   *clients.ProtoClient
	secret   string
	notifier proto.NotifyServer

	mux *http.ServeMux
}

func New(client *clients.ProtoClient, secret string, notifier proto.NotifyServer, routePrefix string) http.Handler {
	g := &_GitHub{
		client:   client,
		secret:   secret,
		notifier: notifier,
		mux:      http.NewServeMux(),
	}
	g.register()
	return http.StripPrefix(routePrefix, g.mux)
}

func (g *_GitHub) register() {
	g.mux.HandleFunc(`POST /`, g.onRecv)
}

func (g *_GitHub) onRecv(w http.ResponseWriter, r *http.Request) {
	sum := strings.TrimPrefix(r.Header.Get(`X-Hub-Signature-256`), `sha256=`)
	payload, err := decode(r.Body, g.secret, sum)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if w := payload.WorkflowRun; w.Status == `completed` {
		switch w.Conclusion {
		case `success`:
			_, err := g.client.Management.ScheduleUpdate(
				auth.SystemForGateway(context.Background()),
				&proto.ScheduleUpdateRequest{},
			)
			if err != nil {
				g.notify(`持续集成`, fmt.Sprintf(`启动计划任务失败：%v`, err))
				return
			}
		default:
			g.notify(`持续集成`, fmt.Sprintf("结果未知：%s", w.Conclusion))
		}
	}
}

func (g *_GitHub) notify(subject, body string) {
	g.notifier.SendInstant(
		auth.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Title: subject,
			Body:  body,
			Level: proto.SendInstantRequest_Active,
			Group: `系统状态`,
		},
	)
}
