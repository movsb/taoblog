package gateway

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	proto_docs "github.com/movsb/taoblog/protocols/docs"
	"github.com/movsb/taoblog/protocols/go/handy"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/modules/renderers/plantuml"
	"github.com/movsb/taoblog/service/modules/webhooks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"nhooyr.io/websocket"
)

//go:embed FEATURES.md
var featuresMd []byte
var featuresTime = time.Now()

type Gateway struct {
	mux     *http.ServeMux
	service *service.Service
	auther  *auth.Auth

	instantNotifier notify.InstantNotifier
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux, instantNotifier notify.InstantNotifier) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,

		instantNotifier: instantNotifier,
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
	)

	mux.Handle(`/v3/`, mux2)

	if err := g.register(context.TODO(), mux, mux2); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) register(ctx context.Context, mux *http.ServeMux, mux2 *runtime.ServeMux) error {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100<<20),
			grpc.MaxCallSendMsgSize(100<<20),
		),
	}

	proto.RegisterUtilsHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)
	proto.RegisterTaoBlogHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)
	proto.RegisterSearchHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)

	mux2.HandlePath("GET", `/v3/api`, serveProtoDocsFile(`index.html`))
	mux2.HandlePath("GET", `/v3/api/swagger`, serveProtoDocsFile(`taoblog.swagger.json`))
	mux2.HandlePath(`GET`, `/v3/features/{theme}`, features)

	mux2.HandlePath(`GET`, `/v3/avatar/{id}`, g.getAvatar)

	mux2.HandlePath(`POST`, `/v3/webhooks/github`, g.githubWebhook())

	mux.Handle(`GET /v3/posts/{id}/files`, g.createFileManager(`post`))

	return nil
}

func (g *Gateway) mimicGateway(r *http.Request) context.Context {
	md := metadata.Pairs()
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
		client := proto.NewManagementClient(conn)
		fs, err := client.FileSystem(g.mimicGateway(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		initReq := proto.FileSystemRequest_InitRequest{
			For: &proto.FileSystemRequest_InitRequest_Post_{
				Post: &proto.FileSystemRequest_InitRequest_Post{
					Id: id,
				},
			},
		}
		if err := fs.Send(&proto.FileSystemRequest{Init: &initReq}); err != nil {
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
			g.instantNotifier.InstantNotify("博客更新", content)
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
	in := &handy.GetAvatarRequest{
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

var reFeaturesPlantUML = regexp.MustCompile("```plantuml((?sU).+)```")

func features(w http.ResponseWriter, r *http.Request, params map[string]string) {
	matches := reFeaturesPlantUML.FindSubmatch(featuresMd)
	if len(matches) == 2 {
		compressed, err := plantuml.Compress(matches[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		darkMode := false
		if params[`theme`] == `dark` {
			darkMode = true
		}
		content, err := plantuml.Fetch(r.Context(), `https://www.plantuml.com/plantuml`, `svg`, compressed, darkMode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add(`Content-Type`, `image/svg+xml`)
		http.ServeContent(w, r, `features.svg`, featuresTime, bytes.NewReader(content))
		return
	}
}
