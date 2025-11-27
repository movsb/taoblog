package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	stdRuntime "runtime"
	"slices"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

var jsonPB = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	},
}

func TestFiles(t *testing.T) {
	r := Serve(t.Context())
	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 123`,
	}))

	// 上传文件
	body := bytes.NewBuffer(nil)
	parts := multipart.NewWriter(body)

	utils.Must(parts.WriteField(
		`spec`, string(utils.Must1(jsonPB.Marshal(&proto.FileSpec{
			Path: `blank.png.xxx`,
			Size: 4637,
			Time: 0,

			Meta: &proto.FileSpec_Meta{
				Width:  60,
				Height: 60,
				Source: &proto.FileSpec_Meta_Source{
					Format:  proto.FileSpec_Meta_Source_Markdown,
					Caption: `123`,
				},
			},
		}))),
	))

	fs := os.DirFS(`testdata`)
	fp := utils.Must1(fs.Open(`blank.png`))
	defer fp.Close()
	fw := utils.Must1(parts.CreateFormFile(`data`, `blank.png`))
	utils.Must1(io.Copy(fw, fp))
	utils.Must(parts.Close())

	req := utils.Must1(http.NewRequestWithContext(
		r.user1, http.MethodPost,
		r.server.JoinPath(`/v3/posts`, fmt.Sprint(p.Id), `/files`),
		bytes.NewBuffer(body.Bytes()),
	))
	req.Header.Set(`Content-Type`, parts.FormDataContentType())
	r.addAuth(req, r.user1ID)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	defer rsp.Body.Close()
	rspBody := strings.TrimSpace(string(utils.Must1(io.ReadAll(rsp.Body))))
	if rspBody != `{"spec":{"path":"blank.png.xxx","size":4637,"type":"application/octet-stream","meta":{"width":60,"height":60,"source":{"format":2,"caption":"123"}}}}` {
		t.Error(`返回不正确`)
		t.Log(rspBody)
	}
}

func createFile(t *testing.T, r *R, pid int64, spec *proto.FileSpec, data []byte) {
	// 上传文件
	body := bytes.NewBuffer(nil)
	parts := multipart.NewWriter(body)

	utils.Must(parts.WriteField(
		`spec`, string(utils.Must1(jsonPB.Marshal(spec))),
	))
	_ = utils.Must1(parts.CreateFormFile(`data`, spec.Path))
	utils.Must(parts.Close())

	endpoint := r.server.JoinPath(`/v3/posts`, fmt.Sprint(pid), `files`)

	req := utils.Must1(http.NewRequestWithContext(
		r.user1, http.MethodPost,
		endpoint, bytes.NewBuffer(body.Bytes()),
	))
	req.Header.Set(`Content-Type`, parts.FormDataContentType())
	r.addAuth(req, r.user1ID)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		t.Fatalf(`文件上传错误：status=%s`, rsp.Status)
	}
}

func TestListFilesOptions(t *testing.T) {
	r := Serve(t.Context())

	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 123`,
	}))

	createFile(t, r, p.Id, &proto.FileSpec{Path: `1.doc`}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{Path: `1.avif`}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{Path: `1.webm`}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{Path: `1.drawio`}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{Path: `1.drawio.svg`, ParentPath: `1.drawio`}, nil)

	expect := func(t *testing.T, req *proto.ListPostFilesRequest, want []string) {
		req.PostId = int32(p.Id)
		files := utils.Must1(r.client.Blog.ListPostFiles(r.user1, req)).Files
		mapped := utils.Map(files, func(spec *proto.FileSpec) string { return spec.Path })
		slices.Sort(want)
		slices.Sort(mapped)
		if !reflect.DeepEqual(want, mapped) {
			_, file, line, _ := stdRuntime.Caller(1)
			t.Errorf(`文件列表不对(%s:%d): %v, %v`, file, line, want, mapped)
		}
	}

	expect(t, &proto.ListPostFilesRequest{}, []string{`1.doc`, `1.avif`, `1.drawio`})
	expect(t, &proto.ListPostFilesRequest{WithGenerated: true}, []string{`1.doc`, `1.avif`, `1.drawio`, `1.drawio.svg`})
	expect(t, &proto.ListPostFilesRequest{WithLivePhotoVideo: true}, []string{`1.doc`, `1.avif`, `1.webm`, `1.drawio`})
}

func TestCreateEmptyFiles(t *testing.T) {
	r := Serve(t.Context())

	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 123`,
	}))

	createFile(t, r, p.Id, &proto.FileSpec{
		Path: `1.txt`,
	}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{
		Path: `2.txt`,
	}, nil)
}

func TestCreateGeneratedFiles(t *testing.T) {
	r := Serve(t.Context())

	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 123`,
	}))

	createFile(t, r, p.Id, &proto.FileSpec{
		Path: `1.tldraw`,
	}, nil)

	createFile(t, r, p.Id, &proto.FileSpec{
		Path:       `1.tldraw.light.svg`,
		ParentPath: `1.tldraw`,
	}, nil)

	createFile(t, r, p.Id, &proto.FileSpec{
		Path:       `1.tldraw.dark.svg`,
		ParentPath: `1.tldraw`,
	}, nil)

	list := utils.Must1(r.client.Blog.ListPostFiles(r.user1, &proto.ListPostFilesRequest{PostId: int32(p.Id)})).Files
	listNames := utils.Map(list, func(f *proto.FileSpec) string { return f.Path })
	if len(listNames) != 1 || listNames[0] != `1.tldraw` {
		t.Fatal(`列表不正确`, listNames)
	}

	// 删除主文件
	utils.Must1(r.client.Blog.DeletePostFile(r.user1, &proto.DeletePostFileRequest{
		PostId: int32(p.Id),
		Path:   `1.tldraw`,
	}))

	list = utils.Must1(r.client.Blog.ListPostFiles(r.user1, &proto.ListPostFilesRequest{PostId: int32(p.Id)})).Files
	if len(list) != 0 {
		t.Fatal(`列表不正确`, list)
	}
}

func TestFilePerm(t *testing.T) {
	r := Serve(t.Context())
	p := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 123`,
	}))

	createFile(t, r, p.Id, &proto.FileSpec{
		Path: `1.txt`,
	}, nil)
	createFile(t, r, p.Id, &proto.FileSpec{
		Path: `_x.txt`,
	}, nil)

	// 管理员及自己可以访问。
	for _, user := range []context.Context{r.system, r.user1} {
		_, err := r.client.Blog.ListPostFiles(user, &proto.ListPostFilesRequest{PostId: int32(p.Id)})
		if err != nil {
			t.Fatal(`列举文件应该成功`, user, err)
		}
	}
	// 其它用户不可以访问（私密文章）。
	for _, user := range []context.Context{r.admin, r.guest, r.user2} {
		_, err := r.client.Blog.ListPostFiles(user, &proto.ListPostFilesRequest{PostId: int32(p.Id)})
		if err == nil {
			t.Fatal(`列举文件不应该成功`, user)
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
			t.Fatal(`列举文件状态错误`, err)
		}
	}

	// 列举文件暂时仍然会列出不可访问的文件。
	list := utils.Must1(r.client.Blog.ListPostFiles(r.user1, &proto.ListPostFilesRequest{PostId: int32(p.Id)})).Files
	if len(list) != 2 {
		t.Fatal(`文件列表不正确`, list)
	}

	expect := func(t *testing.T, userID int, file string, wantStatus int) {
		req := utils.Must1(http.NewRequest(http.MethodGet, r.server.JoinPath(fmt.Sprint(p.Id), file), nil))
		if userID > 0 {
			r.addAuth(req, int64(userID))
		}
		rsp := utils.Must1(http.DefaultClient.Do(req))
		defer rsp.Body.Close()
		if rsp.StatusCode != wantStatus {
			// io.Copy(os.Stderr, rsp.Body)
			_, file, line, _ := stdRuntime.Caller(1)
			t.Fatal(`文件访问失败`, rsp.Status, file, line)
		}
	}

	// 系统可以访问私有文章，也可以访问特殊文件。
	expect(t, auth.SystemID, `1.txt`, 200)
	expect(t, auth.SystemID, `_x.txt`, 200)

	// 本人可以访问特殊文件。
	expect(t, int(r.user1ID), `1.txt`, 200)
	expect(t, int(r.user1ID), `_x.txt`, 200)

	// 因无法访问私有文章，所以均 404
	expect(t, auth.AdminID, `1.txt`, 404)
	expect(t, auth.AdminID, `_x.txt`, 404)
	expect(t, 0, `1.txt`, 404)
	expect(t, 0, `_x.txt`, 404)

	expect(t, int(r.user2ID), `1.txt`, 404)
	expect(t, int(r.user2ID), `_x.txt`, 404)

	utils.Must1(r.client.Blog.SetPostStatus(r.system, &proto.SetPostStatusRequest{
		Id:     p.Id,
		Status: models.PostStatusPartial,
	}))
	utils.Must1(r.client.Blog.SetPostACL(r.system, &proto.SetPostACLRequest{
		PostId: p.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user2ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	expect(t, int(r.user2ID), `1.txt`, 200)
	expect(t, int(r.user2ID), `_x.txt`, 403)
}
