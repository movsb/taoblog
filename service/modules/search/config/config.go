package search_config

type Config struct {
	InMemory bool        `yaml:"in_memory"`
	Paths    PathsConfig `yaml:"paths"`
}

func DefaultConfig() Config {
	return Config{
		InMemory: true,
		Paths:    DefaultPathsConfig(),
	}
}

type PathsConfig struct {
	// Where to save index data.
	Data string `yaml:"data"`
}

func DefaultPathsConfig() PathsConfig {
	return PathsConfig{
		Data: `.index`,
	}
}
