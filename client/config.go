package main

// HostConfig is a per host config.
type HostConfig struct {
	API    string `yaml:"api"`
	Verify bool   `yaml:"verify"`
	Token  string `yaml:"token"`
}
