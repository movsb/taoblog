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
		panic(err)
	}
	defer fp.Close()
	return Load(fp)
}

// Load ...
func Load(r io.Reader) *Config {
	c := DefaultConfig()
	err := yaml.NewDecoder(r).Decode(&c)
	if err != nil {
		panic(err)
	}
	return &c
}
