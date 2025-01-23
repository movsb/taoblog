package config

// AuthConfig ...
type AuthConfig struct {
	// Github GithubAuthConfig `yaml:"github"`
	// Google GoogleAuthConfig `yaml:"google"`

	AdminName   string
	AdminEmails []string
}

// DefaultAuthConfig ...
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		// Github: DefaultGithubAuthConfig(),
		// Google: DefaultGoogleAuthConfig(),
	}
}

// GithubAuthConfig ...
type GithubAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// DefaultGithubAuthConfig ...
func DefaultGithubAuthConfig() GithubAuthConfig {
	return GithubAuthConfig{}
}

// GoogleAuthConfig ...
type GoogleAuthConfig struct {
	ClientID string `yaml:"client_id"`
}

// DefaultGoogleAuthConfig ...
func DefaultGoogleAuthConfig() GoogleAuthConfig {
	return GoogleAuthConfig{}
}
