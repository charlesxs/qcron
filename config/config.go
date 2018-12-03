package config

import "github.com/BurntSushi/toml"

type cron struct {
	Listen string
	Nodes []string
}

type CronConfig struct {
	Cron cron
}


func NewConfig(confPath string) (*CronConfig, error) {
	var config CronConfig
	if _, err := toml.DecodeFile(confPath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
