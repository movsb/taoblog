package utils

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"
)

// 类似 RSS 这种总是应该只输出公开文章，完全不用管当前是否登录。
func StripCredentialsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(`Cookie`)
		r.Header.Del(`Authorization`)
		h.ServeHTTP(w, r)
	})
}

type HTTPMux interface {
	Handle(pattern string, handler http.Handler)
}

func HTTPError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf(`%d %s`, code, http.StatusText(code)), code)
}

func ServeFSWithModTime(w http.ResponseWriter, r *http.Request, fs fs.FS, t time.Time, file string) {
	fp, err := http.FS(fs).Open(file)
	if err == nil {
		defer fp.Close()
		http.ServeContent(w, r, file, t, fp)
		return
	}
	// 仅用于标准的错误处理，文件已经在上面处理过了。
	http.ServeFileFS(w, r, fs, file[1:])
}

// 总是使用 base64 编码/带 content-type 的 data url
// data:image/svg+xml;base64,AAA
type DataURL struct {
	Type string
	Data []byte
}

func ParseDataURL(u string) (_ *DataURL, outErr error) {
	defer CatchAsError(&outErr)
	outURL := &DataURL{}
	url := Must1(urlpkg.Parse(u))
	if url.Scheme == `data` {
		ty, after, found := strings.Cut(url.Opaque, `;`)
		if found {
			outURL.Type = ty
			enc, data, found := strings.Cut(after, `,`)
			if found && enc == `base64` {
				bin, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					return nil, err
				}
				outURL.Data = bin
				return outURL, nil
			}
		}
	}
	return nil, fmt.Errorf(`cannot parse as data url: %s`, u)
}

func CreateDataURL(d []byte) *DataURL {
	a, _, _ := strings.Cut(http.DetectContentType(d), `;`)
	return &DataURL{
		Type: a,
		Data: d,
	}
}

func (d DataURL) String() string {
	s := d.Type + `;` + `base64,` + base64.StdEncoding.EncodeToString(d.Data)
	return (&urlpkg.URL{Scheme: `data`, Opaque: s}).String()
}
