package client

import (
	"errors"
	"log"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/status"
)

// HostConfig is a per host config.
type HostConfig struct {
	Home  string `yaml:"home"`
	Token string `yaml:"token"`
}

func (c *Client) GetConfig(path string) (string, error) {
	rsp, err := c.Management.GetConfig(c.Context(), &proto.GetConfigRequest{
		Path: path,
	})
	if err != nil {
		return ``, err
	}
	return rsp.Yaml, nil
}

func (c *Client) SetConfig(path string, value string) error {
	_, err := c.Management.SetConfig(c.Context(), &proto.SetConfigRequest{
		Path: path,
		Yaml: value,
	})
	if err != nil {
		return errors.New(status.Convert(err).Message())
	}
	return nil
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
