package client

import (
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestParsePostAssets(t *testing.T) {
	tests := []struct {
		Source string
		Assets []string
	}{
		{
			Source: `a  <a href="a.jpg">adf</a> a`,
			Assets: []string{`a.jpg`},
		},
		{
			Source: `a  <A href="%E4%B8%AD%E6%96%87.mp3">adf</a>`,
			Assets: []string{`中文.mp3`},
		},
		{
			Source: `![](a.jpg?s=.5)`,
			Assets: []string{`a.jpg`},
		},
		{
			Source: `![](#a)`,
			Assets: []string{},
		},
		{
			Source: `![](/123/a.avif)`,
			Assets: []string{},
		},
		{
			Source: `![](https://example.com/123/a.avif)`,
			Assets: []string{},
		},
	}
	for _, t1 := range tests {
		assets, err := parsePostAssets(t1.Source)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(t1.Assets) != len(assets) {
			t.Errorf(`assets not equal: %s`, t1.Source)
			yaml.NewEncoder(os.Stdout).Encode(t1.Assets)
			yaml.NewEncoder(os.Stdout).Encode(assets)
			continue
		}
		for i := 0; i < len(t1.Assets); i++ {
			if !strings.EqualFold(t1.Assets[i], assets[i]) {
				t.Errorf(`assets not equal: %s`, t1.Source)
				continue
			}
		}
	}
}
