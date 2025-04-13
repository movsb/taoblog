package assets

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateFile(client *clients.ProtoClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := auth.Context(r.Context())
		id := utils.Must1(strconv.Atoi(r.PathValue(`id`)))
		po, err := client.Blog.GetPost(
			auth.NewContextForRequestAsGateway(r),
			&proto.GetPostRequest{
				Id: int32(id),
			},
		)
		if err != nil {
			if se, ok := status.FromError(err); ok {
				if se.Code() == codes.NotFound {
					utils.HTTPError(w, 404)
					return
				}
			}
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}
		if ac.User.ID != int64(po.UserId) {
			utils.HTTPError(w, http.StatusForbidden)
			return
		}

		fsc, err := client.Management.FileSystem(
			auth.NewContextForRequestAsGateway(r),
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer fsc.CloseSend()

		utils.Must(fsc.Send(&proto.FileSystemRequest{
			Init: &proto.FileSystemRequest_InitRequest{
				For: &proto.FileSystemRequest_InitRequest_Post_{
					Post: &proto.FileSystemRequest_InitRequest_Post{
						Id: int64(id),
					},
				},
			},
		}))
		utils.Must1(fsc.Recv())

		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

		specValue := r.FormValue(`spec`)
		var spec proto.FileSpec
		// NOTE 普通 json
		dec := json.NewDecoder(strings.NewReader(specValue))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&spec); err != nil {
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}

		data, _, err := r.FormFile(`data`)
		if err != nil {
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}
		defer data.Close()

		all, err := io.ReadAll(data)
		if err != nil {
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}

		utils.Must(fsc.Send(&proto.FileSystemRequest{
			Request: &proto.FileSystemRequest_WriteFile{
				WriteFile: &proto.FileSystemRequest_WriteFileRequest{
					Spec: &spec,
					Data: all,
				},
			},
		}))
		utils.Must1(fsc.Recv())
	})
}
