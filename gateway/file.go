package gateway

import (
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
)

func (g *Gateway) GetFile(w http.ResponseWriter, req *http.Request, params map[string]string) {
	postID := utils.MustToInt64(params["post_id"])
	file := params["file"]
	fp := g.service.GetFile(postID, file)
	http.ServeFile(w, req, fp)
}

func (g *Gateway) CreateFile(w http.ResponseWriter, req *http.Request, params map[string]string) {
	postID := utils.MustToInt64(params["post_id"])
	file := params["file"]
	if err := g.service.CreateFile(postID, file, req.Body); err != nil {
		panic(err)
	}
	w.WriteHeader(200)
}

func (g *Gateway) DeleteFile(w http.ResponseWriter, req *http.Request, params map[string]string) {
	panic(`not impl`)
}
