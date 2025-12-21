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
	stdRuntime "runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
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
		spec, options, data, fsc, ok := readRequest(client, w, r)
		if !ok {
			return
		}

		spec, err := single(fsc, spec, data, options)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			`spec`: spec,
		})
	})
}

func readRequest(client *clients.ProtoClient, w http.ResponseWriter, r *http.Request) (_ *proto.FileSpec, _ Options, _ io.ReadCloser, _ proto.Management_FileSystemClient, _ bool) {
	ac := user.Context(r.Context())
	id := utils.Must1(strconv.Atoi(r.PathValue(`id`)))
	po, err := client.Blog.GetPost(
		user.ForwardRequestContext(r),
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

	fsc, err := client.Management.FileSystem(user.ForwardRequestContext(r))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

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

	r.Body = http.MaxBytesReader(w, r.Body, 101<<20)

	specValue := r.FormValue(`spec`)
	spec := &proto.FileSpec{}
	if err := jsonPB.Unmarshal([]byte(specValue), spec); err != nil {
		log.Println(err)
		utils.HTTPError(w, http.StatusBadRequest)
		return
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

	return spec, options, data, fsc, true
}

func single(fsc proto.Management_FileSystemClient, spec *proto.FileSpec, data io.ReadCloser, options Options) (*proto.FileSpec, error) {
	if spec.Type == `` {
		spec.Type = mime.TypeByExtension(path.Ext(spec.Path))
		if spec.Type == `` {
			spec.Type = `application/octet-stream`
		}
	}

	if shouldConvertImage(spec.Path) || options.DropGPSTags {
		// 仅在必要的时候把数据写到临时文件，避免不必要的开销。
		tmpInputFile := utils.Must1(os.CreateTemp("", "*"+path.Ext(spec.Path)))
		defer os.Remove(tmpInputFile.Name())
		utils.Must1(io.Copy(tmpInputFile, data))

		nextPath := tmpInputFile.Name()

		if shouldConvertImage(nextPath) {
			var err error
			var newPath string
			utils.LimitExec(
				`convertToAVIF`, &numberOfAvifProcesses, stdRuntime.NumCPU(),
				func() {
					// 跨线程写没问题吧？因为 LimitExec 会阻塞直到完成。
					newPath, nextPath, err = convertToAVIF(spec.Path, nextPath)
				},
			)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			spec.Path = newPath
			spec.Type = mime.TypeByExtension(path.Ext(spec.Path))

			defer os.Remove(nextPath)
		}

		if isImageFile(nextPath) && options.DropGPSTags {
			_ = utils.Must1(DropGPSTags(nextPath))
		}

		data.Close()

		fp := utils.Must1(os.Open(nextPath))
		data = fp
		stat := utils.Must1(fp.Stat())
		spec.Size = uint32(stat.Size())
	}

	defer data.Close()

	utils.Must(fsc.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_WriteFile{
			WriteFile: &proto.FileSystemRequest_WriteFileRequest{
				Spec: spec,
				Data: utils.Must1(io.ReadAll(data)),
			},
		},
	}))
	if _, err := fsc.Recv(); err != nil {
		return nil, err
	}

	return spec, nil
}

type Options struct {
	DropGPSTags bool `json:"drop_gps_tags"`
}

var numberOfAvifProcesses atomic.Int32

// TODO:
// 自动加上版权信息（方法一）：
// https://chatgpt.com/share/6880633e-4858-8008-9b01-ba02bdd8c245
// 输出的临时文件需要调用方删除。
func convertToAVIF(rawPath string, inputPath string) (_ string, _ string, outErr error) {
	defer utils.CatchAsError(&outErr)

	newPath, tmpOutputPath := utils.Must2(ConvertToAVIF(context.Background(), rawPath, inputPath, true))

	output := utils.Must1(CopyTags(inputPath, tmpOutputPath))
	// Warning: Error rebuilding maker notes (may be corrupt)
	if strings.Contains(output, `Error rebuilding maker notes`) {
		// utils.Must(DropMakerNotes(tmpOutputPath))
		// 如果拷贝原始文件的元数据失败，可能是因为 MakerNotes 有问题。
		// 直接重新转换，并不拷贝。
		os.Remove(tmpOutputPath)
		newPath, tmpOutputPath = utils.Must2(ConvertToAVIF(context.Background(), rawPath, inputPath, false))
	}

	return newPath, tmpOutputPath, nil
}

// NOTE: 这个接口仅限登录用户使用。
func GetFile(s canonical.FileServer, w http.ResponseWriter, r *http.Request, pid int, path string) {
	_ = user.MustNotBeGuest(r.Context())
	s.ServeFile(w, r, int64(pid), path, true)
}
