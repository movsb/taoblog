package service

import (
	"net/url"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/canonical"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/pingback/xmlrpc"
	"go.uber.org/zap"
)

// Pingback is the pingback service handler.
func (s *Service) Pingback(w xmlrpc.ResponseWriter, source string, target string, title string) {
	zap.L().Info(`pingback`, zap.String(`title`, title), zap.String("source", source), zap.String("target", target))
	targetURL, err := url.Parse(target)
	if err != nil {
		zap.L().Info(`pingback: invalid target`, zap.String(`target`, target))
		w.WriteFault(0, `invalid target`)
		return
	}
	homeURL, _ := url.Parse(s.HomeURL())
	if targetURL.Host != homeURL.Host {
		zap.L().Info(`pingback: invalid host`, zap.String(`target`, target))
		w.WriteFault(0, `invalid host`)
		return
	}
	id, ok := canonical.PostFromPath(targetURL.Path)
	if !ok {
		zap.L().Info(`pingback: not a post url`, zap.String(`target`, target))
		w.WriteFault(0, `invalid post link`)
		return
	}
	s.MustGetPost(id)
	pb := models.Pingback{
		CreatedAt: time.Now().Unix(),
		PostID:    id,
		Title:     title,
		SourceURL: source,
	}
	err = s.tdb.Model(&pb).Create()
	if err != nil {
		if strings.Contains(err.Error(), `UNIQUE constraint`) {
			w.WriteString(`pingback created`)
		} else {
			zap.L().Info(`pingback: create error`, zap.Error(err))
			return
		}
	}
	zap.L().Info(`pingback: created`, zap.String(`target`, target), zap.String(`source`, source))
}

// GetPingbacks ...
func (s *Service) GetPingbacks(id int64) (pingbacks []*models.Pingback) {
	s.tdb.Where(`post_id=?`, id).MustFind(&pingbacks)
	return
}
