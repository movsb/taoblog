package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	ID    int64    `json:"id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type Post struct {
	PostConfig
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
}

type PostStatus struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// InitPost ...
func (c *Client) InitPost() {
	fp, err := os.Open("config.yml")
	if err == nil {
		fp.Close()
		panic("post already initialized, abort")
	}
	config := PostConfig{}
	c.savePostConfig(&config)
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
	resp := c.mustPost("/posts", bytes.NewReader(bys), contentTypeJSON)
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	rp := Post{}
	if err := dec.Decode(&rp); err != nil {
		panic(err)
	}
	p.PostConfig.ID = rp.ID
	c.savePostConfig(&p.PostConfig)
}

func (c *Client) SetPostStatus(status string) {
	config := c.readPostConfig()
	if config.ID == 0 {
		panic("post not yet been created")
	}
	postStatus := &PostStatus{
		ID:     config.ID,
		Status: status,
	}
	bys, err := json.Marshal(postStatus)
	if err != nil {
		panic(err)
	}
	resp := c.mustPost(fmt.Sprintf("/posts/%d/status", config.ID), bytes.NewReader(bys), contentTypeJSON)
	defer resp.Body.Close()
}

func (c *Client) UpdatePost() {
	p := &Post{}
	p.PostConfig = *c.readPostConfig()
	if p.ID == 0 {
		panic("post not created, use create instead")
	}
	sourceType, source := readSource(".")
	p.SourceType = sourceType
	p.Source = source
	bys, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	resp := c.mustPost(fmt.Sprintf("/posts/%d", p.ID), bytes.NewReader(bys), contentTypeJSON)
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	rp := Post{}
	if err := dec.Decode(&rp); err != nil {
		panic(err)
	}
}

// UploadPostFiles ...
func (c *Client) UploadPostFiles() {
	config := c.readPostConfig()
	if config.ID <= 0 {
		panic("post not posted, post it first.")
	}
	includedExts := map[string]bool{
		".jpg": true,
		".png": true,
		".gif": true,
		".zip": true,
		".mp4": true,
	}
	root := "."
	list, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}
	postFiles := []os.FileInfo{}
	fmt.Println("Scanning files ...")
	for _, file := range list {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if _, ok := includedExts[ext]; !ok {
			continue
		}
		postFiles = append(postFiles, file)
		fmt.Println("  +", file.Name())
	}
	fmt.Println("Uploading files ...")
	for _, file := range postFiles {
		fmt.Println("  +", file.Name())
		var err error
		fp, err := os.Open(filepath.Join(root, file.Name()))
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		path := fmt.Sprintf("/posts/%d/files/%s", config.ID, file.Name())
		resp := c.mustPost(path, fp, contentTypeBinary)
		_ = resp
	}
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
