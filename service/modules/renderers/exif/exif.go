package exif

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
)

type Exif struct {
	fs   gold_utils.WebFileSystem
	task *Task
	id   int

	filter gold_utils.NodeFilter
}

type Option func(*Exif)

func WithNodeFilter(f gold_utils.NodeFilter) Option {
	return func(e *Exif) {
		e.filter = f
	}
}

// task: 用于取缓存的，可以为空。
// id: 文章/评论的编号，用于后台成功获取信息后强制使缓存失效以使用元数据信息。
func New(fs gold_utils.WebFileSystem, task *Task, id int, options ...Option) *Exif {
	e := &Exif{
		fs:   fs,
		task: task,
		id:   id,
	}
	for _, opt := range options {
		opt(e)
	}
	return e
}

func (m *Exif) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).FilterFunction(func(i int, s *goquery.Selection) bool {
		if m.filter == nil {
			return true
		}
		return m.filter(s)
	}).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr(`src`, ``)
		if url == "" {
			return
		}
		fp, err := m.fs.OpenURL(url)
		if err != nil {
			log.Println(err)
			return
		}
		// get 会负责关闭文件。
		// defer fp.Close()
		if m.task == nil {
			return
		}

		if metadata := m.task.get(m.id, url, fp); metadata != "" {
			s.SetAttr(`data-metadata`, metadata)
		}
	})
	return nil
}
