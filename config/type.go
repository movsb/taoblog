package config

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
	Metrics     MetricsConfig     `yaml:"metrics"`
	Site        SiteConfig        `yaml:"site"`
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
		Metrics:     DefaultMetricsConfig(),
		Site:        DefaultSiteConfig(),
	}
}

// ServerConfig ...
type ServerConfig struct {
	Listen string             `yaml:"listen"`
	Mailer MailerServerConfig `yaml:"mailer"`
}

// DefaultServerConfig ...
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Listen: `0.0.0.0:2564`,
		Mailer: DefaultMaiMailerServerConfig(),
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
	Path   string `yaml:"path"`
	Mirror string `yaml:"mirror"`
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
	Home        string `yaml:"home"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// DefaultBlogConfig ...
func DefaultBlogConfig() BlogConfig {
	return BlogConfig{
		Home:        `localhost`,
		Name:        `未命名`,
		Description: `还没有写描述哦`,
	}
}
