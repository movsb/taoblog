package config

import search_config "github.com/movsb/taoblog/service/modules/search/config"

// Config ...
type Config struct {
	Database    DatabaseConfig       `yaml:"database"`
	Server      ServerConfig         `yaml:"server"`
	Data        DataConfig           `yaml:"data"`
	Maintenance MaintenanceConfig    `yaml:"maintenance"`
	Auth        AuthConfig           `yaml:"auth"`
	Menus       []MenuItem           `yaml:"menus"`
	Site        SiteConfig           `yaml:"site"`
	Comment     CommentConfig        `yaml:"comment"`
	Search      search_config.Config `yaml:"search"`

	originalFilePath string
}

func (c *Config) Save() {
	if c.originalFilePath == "" {
		panic(`empty config file path`)
	}
	SaveFile(c, c.originalFilePath)
}

// DefaultConfig ...
func DefaultConfig() Config {
	return Config{
		Database:    DefaultDatabaseConfig(),
		Server:      DefaultServerConfig(),
		Data:        DefaultDataConfig(),
		Maintenance: DefaultMainMaintenanceConfig(),
		Auth:        DefaultAuthConfig(),
		Menus:       DefaultMenuConfig(),
		Site:        DefaultSiteConfig(),
		Comment:     DefaultCommentConfig(),
		Search:      search_config.DefaultConfig(),
	}
}

// ServerConfig ...
type ServerConfig struct {
	HTTPListen string             `yaml:"http_listen"`
	GRPCListen string             `yaml:"grpc_listen"`
	Mailer     MailerServerConfig `yaml:"mailer"`
}

// DefaultServerConfig ...
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		HTTPListen: `0.0.0.0:2564`,
		GRPCListen: `0.0.0.0:2563`,
		Mailer:     DefaultMaiMailerServerConfig(),
	}
}

// MailerServerConfig ...
type MailerServerConfig struct {
	Server   string `yaml:"server"`
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
}

// DefaultMaiMailerServerConfig ...
func DefaultMaiMailerServerConfig() MailerServerConfig {
	return MailerServerConfig{}
}

// DataConfig ...
type DataConfig struct {
	File FileDataConfig `yaml:"file"`
}

// DefaultDataConfig ...
func DefaultDataConfig() DataConfig {
	return DataConfig{
		File: DefaultFileDataConfig(),
	}
}

// FileDataConfig ...
type FileDataConfig struct {
	Path string `yaml:"path"`
}

// DefaultFileDataConfig ...
func DefaultFileDataConfig() FileDataConfig {
	return FileDataConfig{
		Path: `./files`,
	}
}

// MaintenanceConfig ...
type MaintenanceConfig struct {
	DisableAdmin bool `yaml:"disable_admin"`
	Webhook      struct {
		ReloaderPath string `yaml:"reloader_path"`
		GitHub       struct {
			Secret string `yaml:"secret"`
		} `yaml:"github"`
	} `yaml:"webhook"`
}

// DefaultMainMaintenanceConfig ...
func DefaultMainMaintenanceConfig() MaintenanceConfig {
	c := MaintenanceConfig{
		DisableAdmin: false,
	}
	c.Webhook.ReloaderPath = `/tmp/taoblog-reloader.sock`
	return c
}
