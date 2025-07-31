package client_common

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/service/models"
)

// TODO 增加 html / blocknote
const IndexFileName = `README.md`

const ConfigFileName = `config.yml`

// PostConfig ...
type PostConfig struct {
	ID       int64           `json:"id" yaml:"id"`
	Title    string          `json:"title" yaml:"title"`
	Modified int32           `json:"modified" yaml:"modified"`
	Tags     []string        `json:"tags" yaml:"tags"`
	Metas    models.PostMeta `json:"metas" yaml:"metas"`
	Slug     string          `json:"slug" yaml:"slug,omitempty"`
	Type     string          `json:"type" yaml:"type"`
}

func ReadPostConfig(path string) (*PostConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadPostConfigReader(file)
}
func ReadPostConfigReader(r io.Reader) (*PostConfig, error) {
	config := PostConfig{}
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func SavePostConfig(path string, config *PostConfig) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := yaml.NewEncoder(file)
	if err := enc.Encode(config); err != nil {
		return err
	}
	return nil
}
