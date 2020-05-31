package config

// DatabaseConfig ...
type DatabaseConfig struct {
	Engine string               `yaml:"engine"`
	MySQL  DatabaseMySQLConfig  `yaml:"mysql"`
	SQLite DatabaseSQLiteConfig `yaml:"sqlite"`
}

// DefaultDatabaseConfig ...
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Engine: `sqlite`,
		MySQL:  DefaultDatabaseMySQLConfig(),
		SQLite: DefaultDatabaseSQLiteConfig(),
	}
}

// DatabaseMySQLConfig ...
type DatabaseMySQLConfig struct {
	Endpoint string `yaml:"endpoint"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultDatabaseMySQLConfig ...
func DefaultDatabaseMySQLConfig() DatabaseMySQLConfig {
	return DatabaseMySQLConfig{
		Endpoint: `localhost:3306`,
		Database: `taoblog`,
		Username: `taoblog`,
		Password: `taoblog`,
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
