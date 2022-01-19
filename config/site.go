package config

// SiteConfig ...
type SiteConfig struct {
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
	return SiteConfig{
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
