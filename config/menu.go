package config

// MenuItem ...
type MenuItem struct {
	Name  string     `yaml:"name"`
	Link  string     `yaml:"link"`
	Blank bool       `yaml:"blank"`
	Items []MenuItem `yaml:"items"`
}

// DefaultMenuConfig ...
func DefaultMenuConfig() []MenuItem {
	return []MenuItem{
		{
			Name: `首页`,
			Link: `/`,
		},
		{
			Name: `管理后台`,
			Link: `/admin/`,
		},
	}
}
