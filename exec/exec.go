package exec

import (
	"log"
	"qcron/task"
	"time"
)


func Run()  {
	go func() {
		for {
			now := time.Now().Unix()

			for _, t := range task.Manager.Tasks {

				if t.TaskTime.NextExecTime.Unix() <= now {
					// TODO 确定执行节点, 如果是自己则广播，等待投票
					// 不是自己则查看下一个任务

					// 执行
					err := t.Run()
					if err != nil {
						log.Println(err)
					}

				}
			}

			time.Sleep(time.Millisecond * 500)
		}

	}()
}
