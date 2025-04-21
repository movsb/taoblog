package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
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
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	return c, dec.Decode(c)
}
