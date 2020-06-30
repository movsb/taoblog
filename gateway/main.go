package gateway

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"google.golang.org/grpc"
)

type Gateway struct {
	router  *gin.RouterGroup
	service *service.Service
	auther  *auth.Auth
}

func NewGateway(router *gin.RouterGroup, service *service.Service, auther *auth.Auth, rootRouter gin.IRouter) *Gateway {
	g := &Gateway{
		router:  router,
		service: service,
		auther:  auther,
	}

	g.routePosts()
	g.routeOthers()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			EmitDefaults: true,
		}),
	)
	rootRouter.Any(`/v3/*path`, func(c *gin.Context) {
		mux.ServeHTTP(c.Writer, c.Request)
	})

	if err := runHTTPService(context.TODO(), mux, service); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) routeOthers() {
	g.router.GET("/avatars/:hash", g.GetAvatar)
}

func (g *Gateway) routePosts() {
	c := g.router.Group("/posts")

	// posts
	c.GET("/:name", g.auth, g.GetPost)
	c.GET("/:name/comments", g.listPostComments)
	c.POST("/:name/comments", g.createPostComment)

	// comments
	c.DELETE("/:name/comments/:comment_name", g.auth, g.DeleteComment)

	// files
	c.GET("/:name/files/*file", g.GetFile)
	c.GET("/:name/files", g.auth, g.ListFiles)
	c.POST("/:name/files/*file", g.auth, g.CreateFile)
	c.DELETE("/:name/files/*file", g.auth, g.DeleteFile)

	// for mirror host
	files := g.router.Group("/files")
	files.GET("/:name/*file", g.GetFile)
}

// runHTTPService ...
// TODO auth
func runHTTPService(ctx context.Context, mux *runtime.ServeMux, svc *service.Service) error {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	protocols.RegisterTaoBlogHandlerFromEndpoint(ctx, mux, service.GrpcAddress, opts)

	compile := func(rule string) httprule.Template {
		if compiler, err := httprule.Parse(rule); err != nil {
			panic(err)
		} else {
			return compiler.Compile()
		}
	}

	handle := func(method string, rule string, handler runtime.HandlerFunc) {
		t := compile(rule)
		pattern, err := runtime.NewPattern(1, t.OpCodes, t.Pool, t.Verb)
		if err != nil {
			panic(err)
		}

		mux.Handle(method, pattern, handler)
	}

	handle("GET", `/v3/api`, getAPI)
	handle("GET", `/v3/api/swagger`, getSwagger)

	return nil
}

func getAPI(w http.ResponseWriter, req *http.Request, params map[string]string) {
	http.ServeFile(w, req, `protocols/docs/index.html`)
}

func getSwagger(w http.ResponseWriter, req *http.Request, params map[string]string) {
	http.ServeFile(w, req, `protocols/docs/taoblog.swagger.json`)
}
