package config

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

func LoadFile(path string) (*Config, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return load(fp)
}

func load(r io.Reader) (*Config, error) {
	c := DefaultConfig()
	dec := yaml.NewDecoder(r, yaml.Strict())
	return c, dec.Decode(c)
}
