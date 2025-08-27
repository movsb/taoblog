package assets

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

var jsonPB = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	},
	UnmarshalOptions: protojson.UnmarshalOptions{},
}

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
		if err := jsonPB.Unmarshal([]byte(specValue), &spec); err != nil {
			log.Println(err)
			utils.HTTPError(w, http.StatusBadRequest)
			return
		}
		if spec.Type == `` {
			spec.Type = mime.TypeByExtension(path.Ext(spec.Path))
			if spec.Type == `` {
				spec.Type = `application/octet-stream`
			}
		}

		var options Options
		optionsString := r.FormValue(`options`)
		if optionsString == "" {
			optionsString = "{}"
		}
		dec := json.NewDecoder(strings.NewReader(optionsString))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&options); err != nil {
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

		specPtr := &spec

		if isImageFile(spec.Path) {
			var (
				spec2 *proto.FileSpec
				data2 []byte
				err   error
			)
			utils.LimitExec(
				`convertToAVIF`, &numberOfAvifProcesses, maxNumberOfAvifProcesses,
				func() {
					// 跨线程写没问题吧？
					spec2, data2, err = convertToAVIF(&spec, all, options.DropGPSTags)
				},
			)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), 500)
				return
			}

			specPtr = spec2
			all = data2
		}

		utils.Must(fsc.Send(&proto.FileSystemRequest{
			Request: &proto.FileSystemRequest_WriteFile{
				WriteFile: &proto.FileSystemRequest_WriteFileRequest{
					Spec: specPtr,
					Data: all,
				},
			},
		}))
		utils.Must1(fsc.Recv())

		json.NewEncoder(w).Encode(map[string]any{
			`spec`: specPtr,
		})
	})
}

type Options struct {
	DropGPSTags bool `json:"drop_gps_tags"`
}

const maxNumberOfAvifProcesses = 1

var numberOfAvifProcesses atomic.Int32

// TODO:
// 自动加上版权信息（方法一）：
// https://chatgpt.com/share/6880633e-4858-8008-9b01-ba02bdd8c245
//
// dropGPSTags: 先完成转换之后再 Drop。
func convertToAVIF(spec *proto.FileSpec, data []byte, dropGPSTags bool) (_ *proto.FileSpec, _ []byte, outErr error) {
	defer utils.CatchAsError(&outErr)

	tmpInputFile := utils.Must1(os.CreateTemp("", ""))
	defer os.Remove(tmpInputFile.Name())
	utils.Must1(tmpInputFile.Write(data))
	tmpInputFile.Close()

	newPath, tmpOutputPath := utils.Must2(ConvertToAVIF(context.Background(), spec.Path, tmpInputFile.Name(), true))

	output := utils.Must1(CopyTags(tmpInputFile.Name(), tmpOutputPath))
	// Warning: Error rebuilding maker notes (may be corrupt)
	if strings.Contains(output, `Error rebuilding maker notes`) {
		// utils.Must(DropMakerNotes(tmpOutputPath))
		// 如果拷贝原始文件的元数据失败，可能是因为 MakerNotes 有问题。
		// 直接重新转换，并不拷贝。
		os.Remove(tmpOutputPath)
		newPath, tmpOutputPath = utils.Must2(ConvertToAVIF(context.Background(), spec.Path, tmpInputFile.Name(), false))
	}

	if dropGPSTags {
		utils.Must1(DropGPSTags(tmpOutputPath))
	}

	fpOutput := utils.Must1(os.Open(tmpOutputPath))
	defer os.Remove(tmpOutputPath)
	defer fpOutput.Close()

	info := utils.Must1(fpOutput.Stat())

	specOutput := &proto.FileSpec{
		Path: newPath,
		Mode: spec.Mode,
		Size: uint32(info.Size()),
		Time: spec.Time,
		Type: mime.TypeByExtension(path.Ext(newPath)),
		Meta: spec.Meta, // TODO: 需要转换吗？能拷贝吗？能直接用吗？
	}

	return specOutput, utils.Must1(io.ReadAll(fpOutput)), nil
}

// NOTE: 这个接口仅限登录用户使用。
func GetFile(s canonical.FileServer, w http.ResponseWriter, r *http.Request, pid int, path string) {
	_ = auth.MustNotBeGuest(r.Context())
	s.ServeFile(w, r, int64(pid), path, true)
}
