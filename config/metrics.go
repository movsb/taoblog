package config

// MetricsConfig ...
type MetricsConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultMetricsConfig ...
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Username: `taoblog`,
		Password: `taoblog`,
	}
}
