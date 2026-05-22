package config

type Menus []MenuItem

func (m *Menus) CanSave() {}

type MenuItem struct {
	Name  string     `json:"name" yaml:"name"`
	Link  string     `json:"link" yaml:"link"`
	Blank bool       `json:"blank" yaml:"blank"`
	Items []MenuItem `json:"items" yaml:"items"`
}

func DefaultMenuConfig() []MenuItem {
	return []MenuItem{
		{
			Name: `首页`,
			Link: `/`,
		},
		{
			Name: `后台`,
			Link: `/admin/`,
		},
	}
}
