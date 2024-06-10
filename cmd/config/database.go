package config

// DatabaseConfig ...
type DatabaseConfig struct {
	// 数据库文件路径。
	// 如果不指定，使用内存数据库。
	Path string `yaml:"path"`
}

// DefaultDatabaseConfig ...
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Path: `taoblog.db`,
	}
}
