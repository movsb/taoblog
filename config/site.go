package config

import "time"

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
	Mottoes          []string      `yaml:"mottoes"`
	Since            _Since        `yaml:"since,omitempty"`
	ShowDescription  bool          `yaml:"show_description"`
	ShowStatus       bool          `yaml:"show_status"`
	ShowRelatedPosts bool          `yaml:"show_related_posts"`
	ShowPingbacks    bool          `yaml:"show_pingbacks"`
	Search           SearchConfig  `yaml:"search"`
	Copyright        string        `yaml:"copyright"`
	RSS              RSSConfig     `yaml:"rss"`
	Sitemap          SitemapConfig `yaml:"sitemap"`
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
		ShowPingbacks:    false,
		Search:           DefaultSearchConfig(),
		RSS:              DefaultRSSConfig(),
		Sitemap:          DefaultSitemapConfig(),
	}
}

// SearchConfig ...
type SearchConfig struct {
	Show     bool   `yaml:"show"`
	EngineID string `yaml:"engine_id"`
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
