package client

import (
	"strings"
	"testing"
)

func TestParsePostAssets(t *testing.T) {
	tests := []struct {
		Source string
		Assets []string
	}{
		{
			Source: `a  <a href="a.jpg" /> adf`,
			Assets: []string{`a.jpg`},
		},
		{
			Source: `a  <A href="%E4%B8%AD%E6%96%87.mp3" /> adf`,
			Assets: []string{`中文.mp3`},
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
