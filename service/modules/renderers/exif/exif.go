package exif

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
)

type Exif struct {
	fs gold_utils.WebFileSystem
}

func New(fs gold_utils.WebFileSystem) *Exif {
	return &Exif{
		fs: fs,
	}
}

func (m *Exif) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr(`src`, ``)
		if url == "" {
			return
		}
		fp, err := m.fs.OpenURL(url)
		if err != nil {
			log.Println(err)
			return
		}
		defer fp.Close()
		stat, err := fp.Stat()
		if err != nil {
			log.Println(err)
			return
		}
		info, err := extract(fp)
		if err != nil {
			log.Println(err)
			return
		}
		info.FileName = filepath.Base(stat.Name())
		info.FileSize = utils.ByteCountIEC(stat.Size())
		s.SetAttr(`data-metadata`, string(utils.DropLast1(json.Marshal(info.String()))))
	})
	return nil
}

// TODO 直接传文件？否则文件大小只能读完才知道。
func extract(r io.Reader) (*Metadata, error) {
	cmd := exec.CommandContext(context.TODO(), `exiftool`, `-G`, `-s`, `-json`, `-`)
	cmd.Stdin = r
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var md []*Metadata
	if err := json.Unmarshal(output, &md); err != nil {
		return nil, err
	}
	if len(md) <= 0 {
		return nil, nil
	}
	return md[0], nil
}
