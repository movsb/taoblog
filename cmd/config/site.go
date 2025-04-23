package config

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

type _Since time.Time

func (s *_Since) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	t := (*time.Time)(s)
	r, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	*t = r
	return nil
}

func (s _Since) String() string {
	return time.Time(s).Format(`2006年01月02日`)
}

func (s _Since) Days() int {
	return int(time.Since(time.Time(s)).Hours()) / 24
}

// SiteConfig ...
type SiteConfig struct {
	Home        string `yaml:"home"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Since       _Since `yaml:"since,omitempty"`

	Notify SiteNotifyConfig `json:"notify" yaml:"notify"`

	Sync SiteSyncConfig `yaml:"sync"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	since := _Since(time.Date(2014, time.December, 24, 0, 0, 0, 0, time.Local))
	return SiteConfig{
		Home:        `http://localhost:2564/`,
		Name:        `未命名`,
		Description: ``,
		Since:       since,
		Notify:      DefaultSiteNotifyConfig(),
	}
}

type SiteNotifyConfig struct {
	NewPost bool `json:"new_post" yaml:"new_post"`
}

func DefaultSiteNotifyConfig() SiteNotifyConfig {
	return SiteNotifyConfig{
		NewPost: true,
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
		Template: `<link rel="stylesheet" type="text/css" href="{{.Source}}" />`,
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

func (c ThemeVariablesConfig) AfterSet(paths Segments, obj any) {
	select {
	case c.changed <- struct{}{}:
	default:
	}
}

func (c *ThemeVariablesConfig) Reload() <-chan struct{} {
	return c.changed
}

type SiteSyncConfig struct {
	R2     OSSConfigWithEnabled `yaml:"r2"`
	COS    OSSConfigWithEnabled `yaml:"cos"`
	Aliyun OSSConfigWithEnabled `yaml:"aliyun"`
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
