package config

type DatabaseConfig struct {
	// 文章数据库文件路径。
	// 如果不指定，使用内存数据库。
	Posts string `yaml:"posts"`

	// 文件数据库文件路径。
	// 如果不指定，使用内存数据库。
	Files string `yaml:"files"`
}

func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Posts: `posts.db`,
		Files: `files.db`,
	}
}
