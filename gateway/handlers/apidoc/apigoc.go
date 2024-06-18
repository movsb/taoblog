package apidoc

import (
	"embed"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
	proto_docs "github.com/movsb/taoblog/protocols/docs"
)

//go:embed index.html
var _root embed.FS

type _ApiDoc struct {
	http.Handler
}

func New() http.Handler {
	root := utils.NewOverlayFS(_root, proto_docs.Root)
	h := &_ApiDoc{
		Handler: http.FileServerFS(root),
	}
	return h
}
