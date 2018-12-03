package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"testing"
)

func TestConfig(t *testing.T)  {
	var conf CronConfig
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println("conf: ", conf.Cron)
}