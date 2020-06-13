package config

// SiteConfig ...
type SiteConfig struct {
	ShowStatus       bool               `yaml:"show_status"`
	ShowRelatedPosts bool               `yaml:"show_related_posts"`
	Search           GoogleSearchConfig `yaml:"search"`
	Copyright        string             `yaml:"copyright"`
	RSS              RSSConfig          `yaml:"rss"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		ShowStatus:       false,
		ShowRelatedPosts: false,
		Search:           DefaultGoogleSearchConfig(),
		RSS:              DefaultRSSConfig(),
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
