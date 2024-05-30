package client

import (
	"log"

	proto "github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc/status"
)

// HostConfig is a per host config.
type HostConfig struct {
	API   string `yaml:"api"`
	GRPC  string `yaml:"grpc"`
	Token string `yaml:"token"`
}

func (c *Client) GetConfig(path string) string {
	rsp, err := c.Management.GetConfig(c.Context(), &proto.GetConfigRequest{
		Path: path,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return rsp.Yaml
}

func (c *Client) SetConfig(path string, value string) {
	rsp, err := c.Management.SetConfig(c.Context(), &proto.SetConfigRequest{
		Path: path,
		Yaml: value,
	})
	if err != nil {
		log.Fatalln(status.Convert(err).Message())
	}
	_ = rsp
}

func (c *Client) Restart() {
	rsp, err := c.Management.Restart(c.Context(), &proto.RestartRequest{
		Reason: `客户端命令手动重启。`,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_ = rsp
}
