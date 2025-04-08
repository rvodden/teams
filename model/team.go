package model

type Team struct {
	Name                 string   `yaml:"name"`
	InternalSlackChannel string   `yaml:"internal_slack_channel"`
	Members              []string `yaml:"members"`
}
