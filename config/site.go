package config

// SiteConfig ...
type SiteConfig struct {
	ShowStatus       bool               `yaml:"show_status"`
	ShowRelatedPosts bool               `yaml:"show_related_posts"`
	Search           GoogleSearchConfig `yaml:"search"`
	Copyright        string             `yaml:"copyright"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		ShowStatus:       false,
		ShowRelatedPosts: false,
		Search:           DefaultGoogleSearchConfig(),
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
