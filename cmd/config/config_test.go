package config_test

import (
	"testing"
	"testing/fstest"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
)

func TestLoad(t *testing.T) {
	c := config.DefaultConfig()
	h := config.DefaultConfig().Site.Home
	if c.Site.Name != `未命名` {
		t.Fatal(`error name`)
	}
	if c.Site.Home == `` || c.Site.Home != h {
		t.Fatal(`error home`)
	}
	fs := fstest.MapFS{
		`config.yaml`: {
			Data: []byte(`site:
  name: 原始`),
			Mode: 0644,
		},
		`config_override.yaml`: {
			Data: []byte(`site:
  name: 覆盖`),
			Mode: 0644,
		},
	}
	utils.Must(config.ApplyFromFile(c, fs, `config.yaml`))
	if c.Site.Name != `原始` {
		t.Fatal(`error name after first load`)
	}
	if c.Site.Home != h {
		t.Fatal(`error home after first load`)
	}
	utils.Must(config.ApplyFromFile(c, fs, `config_override.yaml`))
	if c.Site.Name != `覆盖` {
		t.Fatal(`error name after second load`)
	}
	if c.Site.Home != h {
		t.Fatal(`error home after second load`)
	}
}
