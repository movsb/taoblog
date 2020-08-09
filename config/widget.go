package config

// WidgetsConfig ...
type WidgetsConfig struct {
	Reward WidgetRewardConfig `yaml:"reward"`
}

// DefaultWidgetsConfig ...
func DefaultWidgetsConfig() WidgetsConfig {
	return WidgetsConfig{
		Reward: DefaultWidgetRewardConfig(),
	}
}

// WidgetRewardConfig ...
type WidgetRewardConfig struct {
	WeChat string `yaml:"wechat"`
	AliPay string `yaml:"alipay"`
}

// DefaultWidgetRewardConfig ...
func DefaultWidgetRewardConfig() WidgetRewardConfig {
	return WidgetRewardConfig{}
}
