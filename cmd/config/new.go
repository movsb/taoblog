package config

type OthersConfig struct {
	Geo GeoConfig `json:"geo" yaml:"geo"`
}

type GeoConfig struct {
	// TODO 名字写错了。
	GeoDe GaoDe `json:"gaode" yaml:"gaode"`
}

type GaoDe struct {
	Key string `json:"key" yaml:"key"`
}

func (GaoDe) CanSave() {}
