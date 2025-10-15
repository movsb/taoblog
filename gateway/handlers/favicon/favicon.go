package favicon

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

//go:embed favicon.ico
var _default []byte

type Favicon struct {
	Type string
	Mod  time.Time
	Data []byte
}

func NewFavicon() *Favicon {
	return &Favicon{
		Mod:  time.Now(),
		Data: _default,
		Type: http.DetectContentType(_default),
	}
}

func (h *Favicon) SetData(t time.Time, d *utils.DataURL) {
	h.Mod = t
	h.Type = d.Type
	h.Data = d.Data
}

func (h *Favicon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, h.Type)
	http.ServeContent(w, r, `favicon.ico`, h.Mod, bytes.NewReader(h.Data))
}
