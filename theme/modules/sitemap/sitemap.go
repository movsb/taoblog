package sitemap

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

//go:embed sitemap.xml
var tmpl string

// Article ...
type Article struct {
	Link string
}

// Sitemap ...
type Sitemap struct {
	Articles []*Article

	tmpl *template.Template
	svc  proto.TaoBlogServer
	impl service.ToBeImplementedByRpc
}

// New ...
func New(svc proto.TaoBlogServer, impl service.ToBeImplementedByRpc) *Sitemap {
	s := &Sitemap{
		svc:  svc,
		impl: impl,
		tmpl: template.Must(template.New(`sitemap`).Parse(tmpl)),
	}

	return s
}

// ServeHTTP ...
func (s *Sitemap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rsp, err := s.impl.ListAllPostsIds(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info := utils.Must(s.svc.GetInfo(req.Context(), &proto.GetInfoRequest{}))

	rssArticles := make([]*Article, 0, len(rsp))
	for _, article := range rsp {
		rssArticle := Article{
			Link: fmt.Sprintf("%s/%d/", info.Home, article),
		}
		rssArticles = append(rssArticles, &rssArticle)
	}

	cs := *s
	cs.Articles = rssArticles

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprintln(w, `<?xml version="1.0" encoding="UTF-8"?>`)

	if err := cs.tmpl.Execute(w, cs); err != nil {
		panic(err)
	}
}
