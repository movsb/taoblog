package search

import "time"

type Config struct {
	InMemory     bool          `yaml:"in_memory"`
	Paths        PathsConfig   `yaml:"paths"`
	InitialDelay time.Duration `yaml:"initial_delay"`
	ScanInterval time.Duration `yaml:"scan_interval"`
}

func DefaultConfig() Config {
	return Config{
		InMemory:     true,
		Paths:        DefaultPathsConfig(),
		ScanInterval: time.Minute * 10,
		InitialDelay: time.Second * 10,
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
