package config

// SiteConfig ...
type SiteConfig struct {
	Search GoogleSearchConfig `yaml:"search"`
}

// DefaultSiteConfig ...
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Search: DefaultGoogleSearchConfig(),
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
