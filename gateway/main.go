package gateway

import (
	"context"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	proto_docs "github.com/movsb/taoblog/protocols/docs"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/modules/webhooks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"nhooyr.io/websocket"
)

type Gateway struct {
	mux     *http.ServeMux
	service *service.Service
	auther  *auth.Auth
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,
	}

	mux2 := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		),
		// service 的 rpc 请求可能来自 Gateway 或者 client。
		// 添加一个头部以示区分。不同的来源有不同的 auth 附加方法。
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			return metadata.Pairs(`request_from_gateway`, `true`)
		}),
	)

	mux.Handle(`/v3/`, mux2)

	if err := g.register(context.TODO(), mux, mux2); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) register(ctx context.Context, mux *http.ServeMux, mux2 *runtime.ServeMux) error {
	protocols.RegisterTaoBlogHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})
	protocols.RegisterSearchHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})

	mux2.HandlePath("GET", `/v3/api`, serveProtoDocsFile(`index.html`))
	mux2.HandlePath("GET", `/v3/api/swagger`, serveProtoDocsFile(`taoblog.swagger.json`))

	mux2.HandlePath(`GET`, `/v3/avatar/{id}`, g.getAvatar)

	mux2.HandlePath(`POST`, `/v3/webhooks/github`, g.githubWebhook())

	mux.Handle(`GET /v3/posts/{id}/files`, g.createFileManager(`post`))

	return nil
}

func (g *Gateway) mimicGateway(r *http.Request) context.Context {
	md := metadata.Pairs(`request_from_gateway`, `1`)
	for _, cookie := range r.Header.Values(`cookie`) {
		md.Append(auth.GatewayCookie, cookie)
	}
	for _, userAgent := range r.Header.Values(`user-agent`) {
		md.Append(auth.GatewayUserAgent, userAgent)
	}
	return metadata.NewOutgoingContext(r.Context(), md)
}

func (g *Gateway) createFileManager(kind string) http.Handler {
	if kind != `post` {
		panic(`only for post currently`)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里没鉴权，后端会鉴权。
		// 鉴权了会比较好，可以少打开一个到后端的连接。
		ws, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer ws.CloseNow()
		ws.SetReadLimit(-1)

		id := utils.MustToInt64(r.PathValue(`id`))

		conn, err := grpc.DialContext(r.Context(), g.service.GrpcAddress(), grpc.WithInsecure())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		client := protocols.NewManagementClient(conn)
		fs, err := client.FileSystem(g.mimicGateway(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initReq := protocols.FileSystemRequest_InitRequest{
			For: &protocols.FileSystemRequest_InitRequest_Post_{
				Post: &protocols.FileSystemRequest_InitRequest_Post{
					Id: id,
				},
			},
		}
		if err := fs.Send(&protocols.FileSystemRequest{Init: &initReq}); err != nil {
			log.Println(err)
			return
		}
		initRsp, err := fs.Recv()
		if err != nil {
			log.Println("init failed:", err)
			return
		}
		if initRsp.GetInit() == nil {
			log.Println(`init error`)
			return
		}

		NewFileSystemWrapper().fileServer(r.Context(), ws, fs)
	})
}

func (g *Gateway) githubWebhook() runtime.HandlerFunc {
	h := webhooks.CreateHandler(
		g.service.Config().Maintenance.Webhook.GitHub.Secret,
		g.service.Config().Maintenance.Webhook.ReloaderPath,
		func(content string) {
			tk := g.service.Config().Comment.Push.Chanify.Token
			ch := notify.NewOfficialChanify(tk)
			ch.Send("博客更新", content, true)
		},
	)
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		h(w, req)
	}
}

func (g *Gateway) getAvatar(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ephemeral, err := strconv.Atoi(params[`id`])
	if err != nil {
		panic(err)
	}
	in := &protocols.GetAvatarRequest{
		Ephemeral:       ephemeral,
		IfModifiedSince: req.Header.Get("If-Modified-Since"),
		IfNoneMatch:     req.Header.Get("If-None-Match"),
		SetStatus: func(statusCode int) {
			w.WriteHeader(statusCode)
		},
		SetHeader: func(name string, value string) {
			w.Header().Add(name, value)
		},
		W: w,
	}
	g.service.GetAvatar(in)
}

func serveProtoDocsFile(path string) runtime.HandlerFunc {
	_, name := filepath.Split(path)
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		fp, err := proto_docs.Root.Open(path)
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		stat, err := fp.Stat()
		if err != nil {
			panic(err)
		}
		rs, ok := fp.(io.ReadSeeker)
		if !ok {
			panic(`bad embed file`)
		}
		http.ServeContent(w, req, name, stat.ModTime(), rs)
	}
}
