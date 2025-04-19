package dynamic

import (
	"net/http"
	"net/url"

	"github.com/movsb/taoblog/modules/utils"
)

const (
	Prefix        = `/v3/dynamic`
	PrefixSlashed = Prefix + `/`
)

var (
	BaseURL = utils.Must1(url.Parse(PrefixSlashed))
)

func URL(path string) string {
	return BaseURL.JoinPath(path).String()
}

func New(invalidate func()) http.Handler {
	_invalidate = invalidate
	initAll()
	return &Handler{}
}

// 记录由各渲染扩展/插件动态注册的样式/脚本/资源。
//
// TODO 考虑与 gateway/addons 合并？
type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	initAll()
	roots.ServeHTTP(w, r)
}

// 导出的目的主要是给测试用。
func initAll() {
	onceInits.Do(callInits)
	once.Do(initContents)

	if reloadAll.Load() {
		reloadLock.Lock()
		defer reloadLock.Unlock()
		defer reloadAll.Store(false)
		initContents()
	}
}
