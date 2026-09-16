package config

import "fmt"

type OthersConfig struct {
	Geo     GeoConfig           `json:"geo" yaml:"geo"`
	Twitter OthersTwitterConfig `json:"twitter" yaml:"twitter"`
}

type OthersTwitterConfig struct {
	Enabled           bool   `json:"enabled" yaml:"enabled"`
	ConsumerKey       string `json:"consumer_key" yaml:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret" yaml:"consumer_secret"`
	AccessToken       string `json:"access_token" yaml:"access_token"`
	AccessTokenSecret string `json:"access_token_secret" yaml:"access_token_secret"`
}

func (c OthersTwitterConfig) Valid() bool {
	return c.Enabled && c.ConsumerKey != "" && c.ConsumerSecret != "" && c.AccessToken != "" && c.AccessTokenSecret != ""
}

func (c *OthersTwitterConfig) CanSave() {}

func (c *OthersTwitterConfig) BeforeSet(paths Segments, obj any) error {
	if len(paths) > 0 {
		return fmt.Errorf("Twitter 配置必须整体保存")
	}
	newConfig := obj.(OthersTwitterConfig)
	if newConfig.Enabled && !newConfig.Valid() {
		return fmt.Errorf("启用 Twitter 同步时必须填写全部凭据")
	}
	return nil
}

type GeoConfig struct {
	// TODO 名字写错了。
	GeoDe GaoDe `json:"gaode" yaml:"gaode"`
}

type GaoDe struct {
	Key string `json:"key" yaml:"key"`
}

func (GaoDe) CanSave() {}
