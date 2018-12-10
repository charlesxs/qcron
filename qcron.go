package qcron

import (
	"fmt"
	"log"
	"qcron/config"
	"qcron/libs"
	"qcron/libs/hash"
	"qcron/ndcenter"
	"qcron/task"
	"time"
)

var (
	conf *config.CronConfig
	ndc *ndcenter.NDCenter
)


func Run(configPath string)  {
	c, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("config init error: %s\n", err))
	}

	conf = c
	ndc = &ndcenter.NDCenter{
		CronConfig: conf,
		Ch: hash.NewConsistentHash(conf.Cron.Nodes, 100),
	}

	// ndc server run
	go ndc.ServerRun()

	// init cluster
	ndc.Init()

	// run
	go func() {
		for {
			now := time.Now().Unix()
			for _, t := range task.Manager.Tasks {
				if t.TaskTime.NextExecTime.Unix() <= now {
					// 确认节点
					if ok, count := ndc.Ensure(t.TaskID, 0); !ok {
						continue
					} else {
						if count > 0 {
							newTime, err := libs.TimeParse(t.TimeExpress, time.Now())
							if err != nil {
								fmt.Println("qcron::Run time parse error: ", err)
								continue
							}
							t.TaskTime = newTime
						}
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
