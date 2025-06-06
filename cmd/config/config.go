package config

import (
	"io/fs"

	"github.com/goccy/go-yaml"
)

// 从文件读取配置并应用到目标配置。
func ApplyFromFile(to *Config, fsys fs.FS, path string) error {
	fp, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	return yaml.NewDecoder(fp, yaml.Strict()).Decode(to)
}
