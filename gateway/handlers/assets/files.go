package assets

import (
	"log"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc"
	"nhooyr.io/websocket"
)

func New(kind string, addr string) http.Handler {
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

		conn, err := grpc.DialContext(r.Context(), addr, grpc.WithInsecure())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		client := proto.NewManagementClient(conn)
		fs, err := client.FileSystem(r.Context())
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
