package config

// SiteConfig ...
type SiteConfig struct {
	ShowStatus       bool               `yaml:"show_status"`
	ShowRelatedPosts bool               `yaml:"show_related_posts"`
	ShowPingbacks    bool               `yaml:"show_pingbacks"`
	Search           GoogleSearchConfig `yaml:"search"`
	Copyright        string             `yaml:"copyright"`
	RSS              RSSConfig          `yaml:"rss"`
	Sitemap          SitemapConfig      `yaml:"sitemap"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		ShowStatus:       false,
		ShowRelatedPosts: false,
		ShowPingbacks:    false,
		Search:           DefaultGoogleSearchConfig(),
		RSS:              DefaultRSSConfig(),
		Sitemap:          DefaultSitemapConfig(),
	}
}

// GoogleSearchConfig ...
type GoogleSearchConfig struct {
	EngineID string `yaml:"engine_id"`
}

// DefaultGoogleSearchConfig ...
func DefaultGoogleSearchConfig() GoogleSearchConfig {
	return GoogleSearchConfig{}
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
