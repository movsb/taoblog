package config

import "github.com/movsb/taoblog/service/modules/search"

// Config ...
type Config struct {
	Blog        BlogConfig        `yaml:"blog"`
	Database    DatabaseConfig    `yaml:"database"`
	Server      ServerConfig      `yaml:"server"`
	Data        DataConfig        `yaml:"data"`
	Theme       ThemeConfig       `yaml:"theme"`
	Maintenance MaintenanceConfig `yaml:"maintenance"`
	Auth        AuthConfig        `yaml:"auth"`
	Menus       []MenuItem        `yaml:"menus"`
	Site        SiteConfig        `yaml:"site"`
	Comment     CommentConfig     `yaml:"comment"`
	Widgets     WidgetsConfig     `yaml:"widgets"`
	Search      search.Config     `yaml:"search"`
}

// DefaultConfig ...
func DefaultConfig() Config {
	return Config{
		Blog:        DefaultBlogConfig(),
		Database:    DefaultDatabaseConfig(),
		Server:      DefaultServerConfig(),
		Data:        DefaultDataConfig(),
		Theme:       DefaultThemeConfig(),
		Maintenance: DefaultMainMaintenanceConfig(),
		Auth:        DefaultAuthConfig(),
		Menus:       DefaultMenuConfig(),
		Site:        DefaultSiteConfig(),
		Comment:     DefaultCommentConfig(),
		Widgets:     DefaultWidgetsConfig(),
		Search:      search.DefaultConfig(),
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

// ThemeConfig ...
type ThemeConfig struct {
	Name string `yaml:"name"`
}

// DefaultThemeConfig ...
func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Name: `BLOG`,
	}
}

// MaintenanceConfig ...
type MaintenanceConfig struct {
	SiteClosed   bool `yaml:"site_closed"`
	DisableAdmin bool `yaml:"disable_admin"`
}

// DefaultMainMaintenanceConfig ...
func DefaultMainMaintenanceConfig() MaintenanceConfig {
	return MaintenanceConfig{
		SiteClosed:   false,
		DisableAdmin: false,
	}
}

// BlogConfig ...
type BlogConfig struct {
	Home        string   `yaml:"home"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Mottoes     []string `yaml:"mottoes"`
}

// DefaultBlogConfig ...
func DefaultBlogConfig() BlogConfig {
	return BlogConfig{
		Home:        `http://localhost`,
		Name:        `未命名`,
		Description: ``,
	}
}
