package api

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type JSON struct {
	*runtime.JSONPb
}

func (j *JSON) ContentType(_ any) string {
	return `application/json; charset=utf-8`
}

func New(ctx context.Context, client *clients.ProtoClient, before, after func(mux *runtime.ServeMux)) http.Handler {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&JSON{
				JSONPb: &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: true,
					},
				},
			},
		),
	)

	if before != nil {
		before(mux)
	}

	utils.Must(proto.RegisterUtilsHandlerClient(ctx, mux, client.Utils))
	utils.Must(proto.RegisterTaoBlogHandlerClient(ctx, mux, client.Blog))
	utils.Must(proto.RegisterSearchHandlerClient(ctx, mux, client.Search))
	utils.Must(proto.RegisterNotifyHandlerClient(ctx, mux, client.Notify))
	utils.Must(proto.RegisterAuthHandlerClient(ctx, mux, client.Auth))

	if after != nil {
		after(mux)
	}

	return mux
}
