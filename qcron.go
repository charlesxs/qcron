package qcron

import (
	"fmt"
	"log"
	"qcron/config"
	"qcron/libs/hash"
	"qcron/ndcenter"
	"qcron/task"
	"time"
)

var (
	Config *config.CronConfig
	NDC *ndcenter.NDCenter
)


func Run(configPath string)  {
	c, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("config init error: %s\n", err))
	}

	Config = c
	NDC = &ndcenter.NDCenter{
		CronConfig: Config,
		Ch: hash.NewConsistentHash(Config.Cron.Nodes, 100),
	}

	// ndc server run
	go NDC.ServerRun()

	// init cluster
	NDC.Init()

	// run
	go func() {
		for {
			now := time.Now().Unix()
			for _, t := range task.Manager.Tasks {
				if t.TaskTime.NextExecTime.Unix() <= now {
					// 确认节点
					if ! NDC.Ensure(t.TaskID, 0) {
						continue
					}

					// 执行
					go func() {
						err := t.Run()
						if err != nil {
							log.Println(err)
						}
					}()
				}
			}
			time.Sleep(time.Millisecond * 500)
		}

	}()
}
