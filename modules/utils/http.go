package utils

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	urlpkg "net/url"
	"strings"

	"github.com/movsb/taoblog/modules/version"
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

// 如果文件有时间，使用文件时间（本地环境）；
// 否则使用运行起始时间（线上环境，fix embed 没有时间的问题）。
// file: 带 / 前缀。
// 不适合动态生成的且没有时间的内容，务必附带当前时间。或者不要使用。
func ServeFSWithAutoModTime(w http.ResponseWriter, r *http.Request, fs fs.FS, file string) {
	fp, err := http.FS(fs).Open(file)
	if err == nil {
		defer fp.Close()
		t := version.Time
		if st, err := fp.Stat(); err == nil {
			if mod := st.ModTime(); !mod.IsZero() {
				t = mod
			}
		}
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

func AddHeader(h http.Handler, name, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(name, value)
		h.ServeHTTP(w, r)
	})
}

// 判断文件路径是否是本地路径。
func IsLocalPathURL(u *url.URL) bool {
	return u.Scheme == `` && u.Opaque == `` && u.Host == ``
}
