package client

import (
	"log"
	"os"

	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

// HostConfig is a per host config.
type HostConfig struct {
	Home  string `yaml:"home"`
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

func (c *Client) Update() {
	rsp, err := c.Management.ScheduleUpdate(c.Context(), &proto.ScheduleUpdateRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	_ = rsp
}

func (c *Client) Info() {
	rsp, err := c.Blog.GetInfo(c.Context(), &proto.GetInfoRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	yaml.NewEncoder(os.Stdout).Encode(rsp)
}
