package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// LoadFile ...
func LoadFile(path string) *Config {
	fp, err := os.Open(path)
	if err != nil {
		cfg := DefaultConfig()
		return &cfg
	}
	defer fp.Close()
	c := load(fp)
	return c
}

// Load ...
func load(r io.Reader) *Config {
	c := DefaultConfig()
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	err := dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	return &c
}
