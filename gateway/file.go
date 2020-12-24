package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"

	"github.com/movsb/taoblog/modules/utils"
)

// ListFiles ...
func (g *Gateway) ListFiles(w http.ResponseWriter, req *http.Request, params map[string]string) {
	postID := utils.MustToInt64(params["post_id"])
	list, err := g.service.Store().List(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e := json.NewEncoder(w)
	e.SetIndent(``, `  `)
	e.Encode(list)
}

// GetFile ...
func (g *Gateway) GetFile(w http.ResponseWriter, req *http.Request, params map[string]string) {
	postID := utils.MustToInt64(params["post_id"])
	file := params["file"]
	fp, err := g.service.Store().Open(postID, filepath.Clean(file))
	if err != nil {
		// TODO(movsb): format code
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, req, stat.Name(), stat.ModTime(), fp)
}

// CreateFile ...
func (g *Gateway) CreateFile(w http.ResponseWriter, req *http.Request, params map[string]string) {
	postID := utils.MustToInt64(params["post_id"])
	file := params["file"]
	fp, err := g.service.Store().Create(postID, filepath.Clean(file))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fp.Close()
	if _, err := io.Copy(fp, req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}
