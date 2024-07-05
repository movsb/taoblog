package config

import search_config "github.com/movsb/taoblog/service/modules/search/config"

// Config ...
type Config struct {
	Database    DatabaseConfig       `yaml:"database"`
	Server      ServerConfig         `yaml:"server"`
	Data        DataConfig           `yaml:"data"`
	Maintenance MaintenanceConfig    `yaml:"maintenance"`
	Auth        AuthConfig           `yaml:"auth"`
	Menus       Menus                `json:"menus" yaml:"menus"`
	Site        SiteConfig           `yaml:"site"`
	Comment     CommentConfig        `yaml:"comment"`
	Search      search_config.Config `yaml:"search"`
	Others      OthersConfig         `json:"others" yaml:"others"`

	// 尽管站点字体应该由各主题提供，但是为了能跨主题共享字体（减少配置麻烦），
	// 所以我就在这里定义了针对所有站点适用的自定义样式表（或主题）集合。
	Theme ThemeConfig `json:"theme" yaml:"theme"`

	VPS    VpsConfig          `json:"vps" yaml:"vps"`
	Notify NotificationConfig `json:"notify" yaml:"notify"`
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
		Theme:       DefaultThemeConfig(),
	}
}

// ServerConfig ...
type ServerConfig struct {
	HTTPListen string `yaml:"http_listen"`
	GRPCListen string `yaml:"grpc_listen"`
}

// DefaultServerConfig ...
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		HTTPListen: `0.0.0.0:2564`,
		GRPCListen: `0.0.0.0:2563`,
	}
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
	// 如果路径为空，使用内存文件系统。
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
		GitHub struct {
			Secret string `yaml:"secret"`
		} `yaml:"github"`
	} `yaml:"webhook"`
}

// DefaultMainMaintenanceConfig ...
func DefaultMainMaintenanceConfig() MaintenanceConfig {
	c := MaintenanceConfig{
		DisableAdmin: false,
	}
	return c
}
