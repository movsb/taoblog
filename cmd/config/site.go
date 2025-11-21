package config

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/movsb/taoblog/modules/globals"
)

type Since int32

func (Since) CanSave() {}

type SiteConfig struct {
	Home        string `yaml:"home"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	Timezone string `yaml:"timezone"`

	// 初始化时写入数据库，表示建站时间。
	Since Since `json:"since" yaml:"since,omitempty"`

	Sync SiteSyncConfig `yaml:"sync"`
}

func (s *SiteConfig) GetHome() string {
	return s.Home
}
func (s *SiteConfig) GetName() string {
	return s.Name
}
func (s *SiteConfig) GetDescription() string {
	return s.Description
}
func (s *SiteConfig) GetTimezoneLocation() *time.Location {
	return globals.LoadTimezoneOrDefault(s.Timezone, globals.SystemTimezone())
}
func (s *SiteConfig) BeforeSet(paths Segments, obj any) error {
	switch key := paths[0].Key; key {
	case `timezone`:
		value := obj.(string)
		loc, err := time.LoadLocation(value)
		if err != nil {
			return err
		}
		s.Timezone = value
		time.Local = loc
	}
	return nil
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Home:        `http://localhost:2564/`,
		Name:        `未命名`,
		Description: ``,
		Timezone:    `Local`,
	}
}

type ThemeConfig struct {
	Stylesheets ThemeStylesheetsConfig `json:"stylesheets" yaml:"stylesheets"`
	Variables   ThemeVariablesConfig   `json:"variables" yaml:"variables"`
}

func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Stylesheets: DefaultThemeStylesheetsConfig(),
		Variables:   DefaultThemeVariablesConfig(),
	}
}

type ThemeStylesheetsConfig struct {
	Template    string `json:"template" yaml:"template"`
	Stylesheets []struct {
		Source string `json:"source" yaml:"source"`
	} `json:"stylesheets" yaml:"stylesheets"`
}

func (ThemeStylesheetsConfig) CanSave() {}

func DefaultThemeStylesheetsConfig() ThemeStylesheetsConfig {
	return ThemeStylesheetsConfig{
		Template: `<link rel="stylesheet" type="text/css" href="{{.Source}}">`,
	}
}

func (c *ThemeStylesheetsConfig) Render() string {
	t := template.Must(template.New(`stylesheet`).Parse(c.Template))
	w := bytes.NewBuffer(nil)
	for _, ss := range c.Stylesheets {
		t.Execute(w, ss)
		fmt.Fprintln(w)
	}
	return w.String()
}

type ThemeVariablesConfig struct {
	Font struct {
		Family string `json:"family" yaml:"family"`
		Mono   string `json:"mono" yaml:"mono"`
		// font-size: 1.2rem;
		Size string `json:"size" yaml:"size"`
		// font-size-adjust: 0.5;
		Adjust string `json:"adjust" yaml:"adjust"`
	} `json:"font" yaml:"font"`
	Colors struct {
		Accent    string `json:"accent" yaml:"accent"`
		Highlight string `json:"highlight" yaml:"highlight"`
		Selection string `json:"selection" yaml:"selection"`
	} `json:"colors" yaml:"colors"`

	changed chan struct{}
}

func DefaultThemeVariablesConfig() ThemeVariablesConfig {
	return ThemeVariablesConfig{
		changed: make(chan struct{}),
	}
}

func (ThemeVariablesConfig) CanSave() {}

func (c *ThemeVariablesConfig) ClearStruct() {
	d := DefaultThemeVariablesConfig()
	c.Font = d.Font
	c.Colors = d.Colors
	// 保留 changed 通道。
}

func (c ThemeVariablesConfig) AfterSet(paths Segments, obj any) {
	select {
	case c.changed <- struct{}{}:
	default:
		panic(`无法发送 changed 通道。可能是因为没有人监听。`)
	}
}

func (c *ThemeVariablesConfig) Reload() <-chan struct{} {
	return c.changed
}

type SiteSyncConfig struct {
	R2     OSSConfigWithEnabled `yaml:"r2"`
	COS    OSSConfigWithEnabled `yaml:"cos"`
	Aliyun OSSConfigWithEnabled `yaml:"aliyun"`
	Minio  OSSConfigWithEnabled `yaml:"minio"`
}

type OSSConfigWithEnabled struct {
	Enabled   bool `yaml:"enabled"`
	OSSConfig `yaml:",inline"`
}

type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	BucketName      string `yaml:"bucket_name"`
}

func (c *OSSConfig) CanSave() {}
func (c *OSSConfig) BeforeSet(paths Segments, obj any) error {
	return nil
}
