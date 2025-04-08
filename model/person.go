package model

type Person struct {
	Name         string `yaml:"name"`
	SlackChannel string `yaml:"slack_channel"`
	Email        string `yaml:"email"`
}
