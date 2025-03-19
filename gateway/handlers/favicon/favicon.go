package favicon

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"
)

//go:embed favicon.ico
var _default []byte

type Favicon struct {
	t time.Time
	d []byte
}

func NewFavicon() *Favicon {
	return &Favicon{
		t: time.Now(),
		d: _default,
	}
}

func (h *Favicon) SetData(t time.Time, d []byte) {
	h.t = t
	h.d = d
}

func (h *Favicon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, `favicon.ico`, h.t, bytes.NewReader(h.d))
}
