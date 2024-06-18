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
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	since := _Since(time.Date(2014, time.December, 24, 0, 0, 0, 0, time.Local))
	return SiteConfig{
		Home:        `http://localhost:2564/`,
		Name:        `未命名`,
		Description: ``,
		Since:       since,
	}
}

type ThemeConfig struct {
	Stylesheets ThemeStylesheetsConfig `json:"stylesheets" yaml:"stylesheets"`
}

func (ThemeConfig) CanSave() {}

func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Stylesheets: DefaultThemeStylesheetsConfig(),
	}
}

type ThemeStylesheetsConfig struct {
	Template    string `json:"template" yaml:"template"`
	Stylesheets []struct {
		Source string `json:"source" yaml:"source"`
	} `json:"stylesheets" yaml:"stylesheets"`
}

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
