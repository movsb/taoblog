package config

import (
	"time"
)

type _Since time.Time

// func (s _Since) MarshalYAML() (interface{}, error) {
// 	t := (time.Time)(s)
// 	return t.Format(time.RFC3339), nil
// }

func (s *_Since) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
	Home             string        `yaml:"home"`
	Name             string        `yaml:"name"`
	Description      string        `yaml:"description"`
	Since            _Since        `yaml:"since,omitempty"`
	ShowDescription  bool          `yaml:"show_description"`
	ShowStatus       bool          `yaml:"show_status"`
	ShowRelatedPosts bool          `yaml:"show_related_posts"`
	Search           SearchConfig  `yaml:"search"`
	RSS              RSSConfig     `yaml:"rss"`
	Sitemap          SitemapConfig `yaml:"sitemap"`

	// 尽管站点字体应该由各主题提供，但是为了能跨主题共享字体（减少配置麻烦），
	// 所以我就在这里定义了针对所有站点适用的自定义样式表（或主题）集合。
	Theme ThemeConfig `yaml:"theme"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	since := _Since(time.Date(2014, time.December, 24, 0, 0, 0, 0, time.Local))
	return SiteConfig{
		Home:             `http://localhost`,
		Name:             `未命名`,
		Description:      ``,
		Since:            since,
		ShowDescription:  false,
		ShowStatus:       false,
		ShowRelatedPosts: false,
		Search:           DefaultSearchConfig(),
		RSS:              DefaultRSSConfig(),
		Sitemap:          DefaultSitemapConfig(),
		Theme:            DefaultThemeConfig(),
	}
}

// SearchConfig ...
type SearchConfig struct {
	Show bool `yaml:"show"`
}

// DefaultSearchConfig ...
func DefaultSearchConfig() SearchConfig {
	return SearchConfig{}
}

// RSSConfig ...
type RSSConfig struct {
	Enabled      bool `yaml:"enabled"`
	ArticleCount int  `yaml:"article_count"`
}

// DefaultRSSConfig ...
func DefaultRSSConfig() RSSConfig {
	return RSSConfig{
		Enabled:      true,
		ArticleCount: 10,
	}
}

// SitemapConfig ...
type SitemapConfig struct {
	Enabled bool `yaml:"enabled"`
}

// DefaultSitemapConfig ...
func DefaultSitemapConfig() SitemapConfig {
	return SitemapConfig{
		Enabled: true,
	}
}

type ThemeConfig struct {
	Stylesheets ThemeStylesheetsConfig `yaml:"stylesheets"`
}

func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Stylesheets: DefaultThemeStylesheetsConfig(),
	}
}

type ThemeStylesheetsConfig struct {
	Template    string `yaml:"template"`
	Stylesheets []struct {
		Source string `yaml:"source"`
	} `yaml:"stylesheets"`
}

func DefaultThemeStylesheetsConfig() ThemeStylesheetsConfig {
	return ThemeStylesheetsConfig{
		Template: `<link rel="stylesheet" type="text/css" href="{{.Source}}" />`,
	}
}
