package config

import "errors"

type AuthConfig struct {
	Github GithubAuthConfig `yaml:"github"`
	Google GoogleAuthConfig `yaml:"google"`

	AdminName   string
	AdminEmails []string
}

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Github: DefaultGithubAuthConfig(),
		Google: DefaultGoogleAuthConfig(),
	}
}

type GithubAuthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func (GithubAuthConfig) CanSave() {}
func (c *GithubAuthConfig) BeforeSet(paths Segments, obj any) error {
	switch paths.At(0).Key {
	case `client_id`:
		return nil
	case `client_secret`:
		return nil
	}
	return errors.New(`unknown key for github`)
}

func DefaultGithubAuthConfig() GithubAuthConfig {
	return GithubAuthConfig{}
}

// GoogleAuthConfig ...
type GoogleAuthConfig struct {
	ClientID string `yaml:"client_id"`
}

func (GoogleAuthConfig) CanSave() {}
func (c *GoogleAuthConfig) BeforeSet(paths Segments, obj any) error {
	switch paths.At(0).Key {
	case `client_id`:
		return nil
	}
	return errors.New(`unknown key for google`)
}

// DefaultGoogleAuthConfig ...
func DefaultGoogleAuthConfig() GoogleAuthConfig {
	return GoogleAuthConfig{}
}
