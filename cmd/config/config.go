package config

import (
	"io"
	"os"
	"path/filepath"

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

func SaveFile(config *Config, path string) {
	fp, err := os.CreateTemp(filepath.Dir(path), `taoblog-config-*.yaml`)
	if err != nil {
		panic(err)
	}
	if err := yaml.NewEncoder(fp).Encode(config); err != nil {
		fp.Close()
		os.Remove(fp.Name())
		panic(err)
	}
	fp.Close()
	if err := os.Rename(fp.Name(), path); err != nil {
		panic(err)
	}
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

	c.Auth.AdminName = c.Comment.Author
	c.Auth.AdminEmails = c.Comment.Emails

	return &c
}
