package genealogy

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/yuin/goldmark/parser"
)

type Genealogy struct{}

func New() *Genealogy {
	return &Genealogy{}
}

func (e *Genealogy) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	var individuals []*Individual
	d := yaml.NewDecoder(bytes.NewReader(source), yaml.Strict())
	utils.Must(d.Decode(&individuals))

	for _, p := range individuals {
		if p.ID == `` {
			panic(`编号ID不能为空`)
		}
	}

	gen(w, individuals)

	return nil
}
