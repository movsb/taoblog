package e2e_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
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
			Mode: 0600,
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

	endpoint := fmt.Sprintf(`http://%s`, r.server.HTTPAddr())
	u := utils.Must1(url.Parse(endpoint)).JoinPath(`/v3/posts`, fmt.Sprint(p.Id), `/files`)

	req := utils.Must1(http.NewRequestWithContext(
		r.user1, http.MethodPost,
		u.String(), bytes.NewBuffer(body.Bytes()),
	))
	req.Header.Set(`Content-Type`, parts.FormDataContentType())
	r.addAuth(req, r.user1ID)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	defer rsp.Body.Close()
	rspBody := strings.TrimSpace(string(utils.Must1(io.ReadAll(rsp.Body))))
	if rspBody != `{"spec":{"path":"blank.png.xxx","mode":384,"size":4637,"type":"application/octet-stream","meta":{"width":60,"height":60,"source":{"format":2,"caption":"123"}}}}` {
		t.Error(`返回不正确`)
		t.Log(rspBody)
	}
}
