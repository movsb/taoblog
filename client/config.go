package client

// HostConfig is a per host config.
type HostConfig struct {
	API    string `yaml:"api"`
	GRPC   string `yaml:"grpc"`
	Verify bool   `yaml:"verify"`
	Token  string `yaml:"token"`
}
