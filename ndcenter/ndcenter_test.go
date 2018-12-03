package ndcenter

import (
	"fmt"
	"qcron/config"
	"qcron/libs/hash"
	"testing"
)

func TestNDCenter_Ensure(t *testing.T) {
	c, err := config.NewConfig("/Users/charles/shells/go/src/qcron/config/config.toml")
	if err != nil {
		fmt.Println(err)
		return
	}

	ndc := &NDCenter{
		CronConfig: c,
		Ch: hash.NewConsistentHash(c.Cron.Nodes, 100),
	}

	go ndc.ServerRun()
	fmt.Println(ndc.Ensure("MyTask1", 0))
}