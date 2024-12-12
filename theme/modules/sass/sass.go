package sass

import (
	"io/fs"
	"strings"
	"time"

	"github.com/bep/godartsass/v2"
)

func CompileFS(dir fs.FS, main string) (string, error) {
	tr, err := godartsass.Start(godartsass.Options{
		DartSassEmbeddedFilename: `sass`,
		Timeout:                  time.Second * 10,
	})
	if err != nil {
		return ``, err
	}
	defer tr.Close()

	source, err := fs.ReadFile(dir, main)
	if err != nil {
		return ``, err
	}

	result, err := tr.Execute(godartsass.Args{
		Source:         string(source),
		ImportResolver: &_ImportResolver{dir: dir},
	})
	if err != nil {
		return ``, err
	}
	return result.CSS, nil
}

type _ImportResolver struct {
	dir fs.FS
}

func (r *_ImportResolver) CanonicalizeURL(url string) (string, error) {
	if !strings.HasSuffix(url, `.scss`) {
		url += `.scss`
	}
	url = `file:///` + url // 相对的
	return url, nil
}

func (r *_ImportResolver) Load(canonicalizedURL string) (godartsass.Import, error) {
	path := strings.TrimPrefix(canonicalizedURL, `file:///`)
	data, err := fs.ReadFile(r.dir, path)
	if err != nil {
		return godartsass.Import{}, err
	}
	return godartsass.Import{
		Content: string(data),
	}, nil
}
