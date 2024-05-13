package config

// DatabaseConfig ...
type DatabaseConfig struct {
	// 只支持 sqlite3
	Engine string               `yaml:"engine"`
	SQLite DatabaseSQLiteConfig `yaml:"sqlite"`
}

// DefaultDatabaseConfig ...
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Engine: `sqlite`,
		SQLite: DefaultDatabaseSQLiteConfig(),
	}
}

// DatabaseSQLiteConfig ...
type DatabaseSQLiteConfig struct {
	// 数据库文件路径。
	// 如果不指定，使用内存数据库。
	Path string `yaml:"path"`
}

// DefaultDatabaseSQLiteConfig ...
func DefaultDatabaseSQLiteConfig() DatabaseSQLiteConfig {
	return DatabaseSQLiteConfig{
		Path: `taoblog.db`,
	}
}
