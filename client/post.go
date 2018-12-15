package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var sourceNames = []string{
	"README.md",
	"index.md",
	"README.html",
	"index.html",
}

type PostConfig struct {
	ID    int64
	Title string
	Tags  []string
}

type Post struct {
	PostConfig
	SourceType string `json:"source_type"`
	Source     string
}

// CreatePost ...
func (c *Client) CreatePost() {
	p := &Post{}
	p.PostConfig = *c.readPostConfig()
	if p.ID > 0 {
		panic("post already posted, use update instead")
	}
	sourceType, source := readSource(".")
	p.SourceType = sourceType
	p.Source = source
	bys, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	resp := c.mustPost("/posts", bytes.NewReader(bys))
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	rp := Post{}
	if err := dec.Decode(&rp); err != nil {
		panic(err)
	}
	p.PostConfig.ID = rp.ID
	c.savePostConfig(&p.PostConfig)
}

func (c *Client) readPostConfig() *PostConfig {
	config := PostConfig{}
	file, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	dec := yaml.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		panic(err)
	}
	return &config
}

func (c *Client) savePostConfig(config *PostConfig) {
	file, err := os.Create("config.yml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := yaml.NewEncoder(file)
	if err := enc.Encode(config); err != nil {
		panic(err)
	}
}

func readSource(dir string) (string, string) {
	var source string
	var theName string

	for _, name := range sourceNames {
		path := filepath.Join(dir, name)
		bys, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}
		source = string(bys)
		theName = name
		break
	}
	if source == "" {
		panic("source cannot be found")
	}

	if strings.IndexByte(source, '\x08') != -1 {
		panic("source cannot have '\\x08' characters")
	}

	typ := ""
	switch filepath.Ext(theName) {
	case ".md":
		typ = "markdown"
	case ".html":
		typ = "html"
	}

	return typ, source
}
