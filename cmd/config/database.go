package config

// DatabaseConfig ...
type DatabaseConfig struct {
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
	Path string `yaml:"path"`
}

// DefaultDatabaseSQLiteConfig ...
func DefaultDatabaseSQLiteConfig() DatabaseSQLiteConfig {
	return DatabaseSQLiteConfig{
		Path: `taoblog.db`,
	}
}
