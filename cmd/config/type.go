package config

import (
	"fmt"

	"github.com/movsb/taoblog/modules/utils"
	search_config "github.com/movsb/taoblog/service/modules/search/config"
)

// Config ...
type Config struct {
	Database    DatabaseConfig       `yaml:"database"`
	Server      ServerConfig         `yaml:"server"`
	Maintenance MaintenanceConfig    `yaml:"maintenance"`
	Menus       Menus                `json:"menus" yaml:"menus"`
	Site        SiteConfig           `yaml:"site"`
	Search      search_config.Config `yaml:"search"`
	Others      OthersConfig         `json:"others" yaml:"others"`

	// 尽管站点字体应该由各主题提供，但是为了能跨主题共享字体（减少配置麻烦），
	// 所以我就在这里定义了针对所有站点适用的自定义样式表（或主题）集合。
	Theme ThemeConfig `json:"theme" yaml:"theme"`

	Notify NotificationConfig `json:"notify" yaml:"notify"`
}

func DefaultConfig() *Config {
	return &Config{
		Database:    DefaultDatabaseConfig(),
		Server:      DefaultServerConfig(),
		Maintenance: DefaultMainMaintenanceConfig(),
		Menus:       DefaultMenuConfig(),
		Site:        DefaultSiteConfig(),
		Search:      search_config.DefaultConfig(),
		Theme:       DefaultThemeConfig(),
	}
}

func DefaultDemoConfig() *Config {
	c := DefaultConfig()
	c.Database = DatabaseConfig{
		Posts: ``,
		Files: ``,
		Cache: ``,
	}
	return c
}

type ServerConfig struct {
	HTTPListen string `yaml:"http_listen"`
}

// DefaultServerConfig ...
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		HTTPListen: `0.0.0.0:2564`,
	}
}

// MaintenanceConfig ...
type MaintenanceConfig struct {
	Webhook struct {
		GitHub struct {
			Secret string `yaml:"secret"`
		} `yaml:"github"`
	} `yaml:"webhook"`

	Backups struct {
		Sync MaintenanceBackupsSyncConfig     `yaml:"sync"`
		R2   MaintenanceBackupsRemoteR2Config `yaml:"r2"`
	} `yaml:"backups"`
}

func DefaultMainMaintenanceConfig() MaintenanceConfig {
	c := MaintenanceConfig{}
	return c
}

type MaintenanceBackupsSyncConfig struct {
	Enabled bool `yaml:"enabled"`

	URL      string `yaml:"url"`      // Git 仓库地址
	Username string `yaml:"username"` // Git 仓库用户名
	Password string `yaml:"password"` // Git 仓库密码

	Author string `yaml:"author"` // Git 提交作者用户名
	Email  string `yaml:"email"`  // Git 提交作者邮箱
}

func (c *MaintenanceBackupsSyncConfig) CanSave() {}
func (c *MaintenanceBackupsSyncConfig) BeforeSet(paths Segments, obj any) error {
	var (
		new    MaintenanceBackupsSyncConfig
		checks []func() error
	)

	checkAuthor := func() error {
		if new.Author == `` {
			return fmt.Errorf(`git 提交作者者用户名不能为空。`)
		}
		return nil
	}
	checkEmail := func() error {
		if !utils.IsEmail(new.Email) {
			return fmt.Errorf(`git 提交作者者邮箱格式不正确。`)
		}
		return nil
	}
	checkURL := func() error {
		if !utils.IsURL(new.URL, false) {
			return fmt.Errorf(`git 仓库地址不正确`)
		}
		return nil
	}
	checkUsername := func() error {
		if new.Username == `` {
			return fmt.Errorf(`git 仓库用户名格式不正确。`)
		}
		return nil
	}
	checkPassword := func() error {
		if new.Password == `` {
			return fmt.Errorf(`git 仓库密码格式不正确。`)
		}
		return nil
	}

	if len(paths) == 0 {
		new = obj.(MaintenanceBackupsSyncConfig)
		checks = append(checks,
			checkAuthor,
			checkEmail,
			checkURL,
			checkUsername,
			checkPassword,
		)
	} else {
		switch paths[0].Key {
		case `enabled`:
			new.Enabled = obj.(bool)
		case `author`:
			new.Author = obj.(string)
			checks = append(checks, checkAuthor)
		case `email`:
			new.Email = obj.(string)
			checks = append(checks, checkEmail)
		case `url`:
			new.URL = obj.(string)
			checks = append(checks, checkURL)
		case `username`:
			new.Username = obj.(string)
			checks = append(checks, checkUsername)
		case `password`:
			new.Password = obj.(string)
			checks = append(checks, checkPassword)
		default:
			return fmt.Errorf(`未知键值：%v`, paths)
		}
	}

	for _, c := range checks {
		if err := c(); err != nil {
			return err
		}
	}

	return nil
}

type MaintenanceBackupsRemoteConfig struct {
}

type MaintenanceBackupsRemoteR2Config struct {
	// 临时放这儿。
	AgeKey string `yaml:"age_key"`

	OSSConfigWithEnabled `yaml:",inline"`
}

func (c *MaintenanceBackupsRemoteR2Config) CanSave() {}
func (c *MaintenanceBackupsRemoteR2Config) BeforeSet(paths Segments, obj any) error {
	return nil
}
