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

type _Protos struct {
	mux *runtime.ServeMux
	http.Handler
}

func New(ctx context.Context, client clients.Client) http.Handler {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		),
	)

	utils.Must(proto.RegisterUtilsHandlerClient(ctx, mux, client))
	utils.Must(proto.RegisterTaoBlogHandlerClient(ctx, mux, client))
	utils.Must(proto.RegisterSearchHandlerClient(ctx, mux, client))

	return &_Protos{
		mux:     mux,
		Handler: mux,
	}
}
