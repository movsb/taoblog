package config

// Config ...
type Config struct {
	Database    DatabaseConfig    `yaml:"database"`
	Server      ServerConfig      `yaml:"server"`
	Data        DataConfig        `yaml:"data"`
	Theme       ThemeConfig       `yaml:"theme"`
	Maintenance MaintenanceConfig `yaml:"maintenance"`
	Auth        AuthConfig        `yaml:"auth"`
}

// DefaultConfig ...
func DefaultConfig() Config {
	return Config{
		Database:    DefaultDatabaseConfig(),
		Server:      DefaultServerConfig(),
		Data:        DefaultDataConfig(),
		Theme:       DefaultThemeConfig(),
		Maintenance: DefaultMainMaintenanceConfig(),
		Auth:        DefaultAuthConfig(),
	}
}

// DatabaseConfig ...
type DatabaseConfig struct {
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultDatabaseConfig ...
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Database: `taoblog`,
		Username: `taoblog`,
		Password: `taoblog`,
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
		Listen: `localhost:2564`,
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
		Name: `blog`,
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

// AuthConfig ...
type AuthConfig struct {
	Key    string           `yaml:"key"`
	Github GithubAuthConfig `yaml:"github"`
}

// DefaultAuthConfig ...
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Github: DefaultGithubAuthConfig(),
	}
}

// GithubAuthConfig ...
type GithubAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	UserID       int64  `yaml:"user_id"`
}

// DefaultGithubAuthConfig ...
func DefaultGithubAuthConfig() GithubAuthConfig {
	return GithubAuthConfig{}
}
