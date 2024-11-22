package scoped_css

import (
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
)

func New(scope string) *ScopedCSS {
	return &ScopedCSS{
		scope: scope,
	}
}

type ScopedCSS struct {
	scope string
}

func (p *ScopedCSS) TransformHtml(doc *goquery.Document) error {
	var err error
	doc.Find(`style`).Each(func(i int, s *goquery.Selection) {
		if er := p.addScope(s); er != nil {
			err = er
		}
	})
	return err
}

func (p *ScopedCSS) addScope(s *goquery.Selection) error {
	if first := s.Nodes[0].FirstChild; first != nil && first.Type == html.TextNode {
		raw := first.Data
		scoped, err := addScope(raw, p.scope)
		if err != nil {
			return err
		}
		first.Data = scoped
	}
	return nil
}

// 可以参考：https://github.com/tdewolff/parse/blob/master/css/parse_test.go
func addScope(raw string, scope string) (string, error) {
	p := css.NewParser(parse.NewInputString(raw), false)
	output := ""
	var lastGrammarType css.GrammarType
	for {
		grammar, _, data := p.Next()
		data = parse.Copy(data)
		// log.Println(grammar, string(data), p.Values())
		if grammar == css.ErrorGrammar {
			if err := p.Err(); err != io.EOF {
				return ``, err
			}
			break
		}
		if grammar == css.AtRuleGrammar || grammar == css.BeginAtRuleGrammar || grammar == css.QualifiedRuleGrammar || grammar == css.BeginRulesetGrammar || grammar == css.DeclarationGrammar || grammar == css.CustomPropertyGrammar {
			if grammar == css.DeclarationGrammar || grammar == css.CustomPropertyGrammar {
				data = append(data, ":"...)
			}
			if grammar == css.QualifiedRuleGrammar || grammar == css.BeginRulesetGrammar {
				if lastGrammarType == 0 || lastGrammarType == css.QualifiedRuleGrammar || lastGrammarType == css.BeginRulesetGrammar {
					data = append(data, scope...)
					data = append(data, ' ')
				}
			}
			for _, val := range p.Values() {
				data = append(data, val.Data...)
			}
			if grammar == css.BeginAtRuleGrammar || grammar == css.BeginRulesetGrammar {
				data = append(data, "{"...)
			} else if grammar == css.AtRuleGrammar || grammar == css.DeclarationGrammar || grammar == css.CustomPropertyGrammar {
				data = append(data, ";"...)
			} else if grammar == css.QualifiedRuleGrammar {
				data = append(data, ","...)
			}
		}
		output += string(data)
		lastGrammarType = grammar
	}
	return output, nil
}
