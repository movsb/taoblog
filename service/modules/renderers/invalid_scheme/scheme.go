package invalid_scheme

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func New() *_InvalidScheme {
	return &_InvalidScheme{}
}

type _InvalidScheme struct{}

func (*_InvalidScheme) TransformHtml(doc *goquery.Document) error {
	var outErr error
	doc.Find(`a`).Each(func(i int, s *goquery.Selection) {
		if u, ok := s.Attr(`href`); ok {
			parsed, err := url.Parse(u)
			if err != nil {
				log.Println(err)
				outErr = err
				return
			}
			switch strings.ToLower(parsed.Scheme) {
			case ``, `http`, `https`, `tel`, `mailto`, `about`:
				return
			default:
				outErr = fmt.Errorf(`不支持的协议：%q`, parsed.Scheme)
			}
		}
	})
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		if u, ok := s.Attr(`src`); ok {
			parsed, err := url.Parse(u)
			if err != nil {
				log.Println(err)
				outErr = err
				return
			}
			switch strings.ToLower(parsed.Scheme) {
			case ``, `http`, `https`, `data`:
				return
			default:
				outErr = fmt.Errorf(`不支持的协议：%q`, parsed.Scheme)
			}
		}
	})
	return outErr
}
