package config

// AuthConfig ...
type AuthConfig struct {
	Key    string           `yaml:"key"`
	Basic  BasicAuthConfig  `yaml:"basic"`
	Github GithubAuthConfig `yaml:"github"`
	Google GoogleAuthConfig `yaml:"google"`
}

// DefaultAuthConfig ...
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Basic:  DefaultBasicBasicAuthConfig(),
		Github: DefaultGithubAuthConfig(),
		Google: DefaultGoogleAuthConfig(),
	}
}

// BasicAuthConfig ...
type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultBasicBasicAuthConfig ...
func DefaultBasicBasicAuthConfig() BasicAuthConfig {
	return BasicAuthConfig{}
}

// GithubAuthConfig ...
type GithubAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	UserID       int64  `yaml:"user_id"`
}

// DefaultGithubAuthConfig ...
func DefaultGithubAuthConfig() GithubAuthConfig {
	return GithubAuthConfig{}
}

// GoogleAuthConfig ...
type GoogleAuthConfig struct {
	ClientID string `yaml:"client_id"`
	UserID   string `yaml:"user_id"`
}

// DefaultGoogleAuthConfig ...
func DefaultGoogleAuthConfig() GoogleAuthConfig {
	return GoogleAuthConfig{}
}
