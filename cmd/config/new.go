package config

import (
	"fmt"
	"regexp"
)

type OthersConfig struct {
	Whois WhoisConfig `json:"whois" yaml:"whois"`
}

type WhoisConfig struct {
	ApiLayer WhoisApiLayerConfig `json:"api_layer" yaml:"api_layer"`
}

type WhoisApiLayerConfig struct {
	Key string `json:"key" yaml:"key"`
}

func (WhoisApiLayerConfig) CanSave() {}

func (c *WhoisApiLayerConfig) BeforeSet(paths Segments, obj any) error {
	switch paths.At(0).Key {
	case `key`:
		val := obj.(string)
		if val == "" || regexp.MustCompile(`^\w+$`).MatchString(val) {
			return nil
		}
		return fmt.Errorf(`key 的值不合法: %q`, val)
	}
	return nil
}
