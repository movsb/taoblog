package assets

import (
	"log"
	"net/http"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"nhooyr.io/websocket"
)

func New(auther *auth.Auth, kind string, client clients.Client) http.Handler {
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

		ctx := auther.NewContextForRequestAsGateway(r)
		fs, err := client.FileSystem(ctx)
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
