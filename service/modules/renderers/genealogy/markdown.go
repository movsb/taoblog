package genealogy

import (
	"bytes"
	"errors"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

type Genealogy struct {
	output      *[]*Individual
	doNotRender bool
}

func New(options ...Option) *Genealogy {
	g := &Genealogy{}

	for _, opt := range options {
		opt(g)
	}

	return g
}

func (e *Genealogy) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	var individuals []*Individual
	d := yaml.NewDecoder(bytes.NewReader(source), yaml.Strict())

	if err := d.Decode(&individuals); err != nil {
		gold_utils.RenderError(w, err)
		return nil
	}

	// TODO: name不能重复，整个文章范围内。
	for _, p := range individuals {
		if p.ID == `` {
			gold_utils.RenderError(w, errors.New(`编号ID不能为空`))
			return nil
		}
	}

	if e.output != nil {
		*e.output = append(*e.output, individuals...)
	}

	if !e.doNotRender {
		gen(w, individuals)
	}

	return nil
}

type Option func(g *Genealogy)

func WithOutput(individuals *[]*Individual) Option {
	return func(g *Genealogy) {
		g.output = individuals
	}
}

func WithoutRender() Option {
	return func(g *Genealogy) {
		g.doNotRender = true
	}
}
