package client

import (
	"log"

	"github.com/movsb/taoblog/protocols"
)

// HostConfig is a per host config.
type HostConfig struct {
	API   string `yaml:"api"`
	GRPC  string `yaml:"grpc"`
	Token string `yaml:"token"`
}

func (c *Client) GetConfig(path string) string {
	rsp, err := c.management.GetConfig(c.token(), &protocols.GetConfigRequest{
		Path: path,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return rsp.Yaml
}

func (c *Client) SetConfig(path string, value string) {
	rsp, err := c.management.SetConfig(c.token(), &protocols.SetConfigRequest{
		Path: path,
		Yaml: value,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_ = rsp
}

func (c *Client) SaveConfig() {
	rsp, err := c.management.SaveConfig(c.token(), &protocols.SaveConfigRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	_ = rsp
}
