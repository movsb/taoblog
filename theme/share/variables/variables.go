package variables

import (
	"bytes"
	"fmt"
	"io/fs"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

type Config struct {
	*config.ThemeVariablesConfig
}

func (c *Config) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(`:root {`)
	if v := c.Colors.Accent; v != `` {
		fmt.Fprintf(buf, `--accent-color: %s;`, v)
	}
	if v := c.Colors.Highlight; v != `` {
		fmt.Fprintf(buf, `--highlight-color: %s;`, v)
	}
	if v := c.Colors.Selection; v != `` {
		fmt.Fprintf(buf, `--selection-background-color: %s;`, v)
	}
	if v := c.Font.Family; v != `` {
		fmt.Fprintf(buf, `--font-normal: %s, %s;`, v, `"Trebuchet MS","Microsoft YaHei",'Noto Sans',sans-serif`)
	}
	if v := c.Font.Mono; v != `` {
		fmt.Fprintf(buf, `--font-mono: %s, %s;`, v, `Monaco,Consolas,monospace`)
	}
	if v := c.Font.Size; v != `` {
		fmt.Fprintf(buf, `--font-size: %s;`, v)
	}
	buf.WriteString(`}`)
	return buf.String()
}

func (c *Config) Open(name string) (fs.File, error) {
	if name == `style.css` {
		return utils.NewStringFile(name, time.Now(), []byte(c.String())), nil
	}
	return nil, fs.ErrNotExist
}

func SetConfig(c *config.ThemeVariablesConfig) {
	cc := &Config{ThemeVariablesConfig: c}
	dynamic.RegisterInit(func() {
		const module = `variables`
		dynamic.WithRoots(module, nil, nil, cc, cc)
		dynamic.WithStyles(module, `style.css`)
	})
}
