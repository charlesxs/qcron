# qcron

**测试demo**

```
$ cat demo.go 
package main

import (
	"flag"
	"github.com/charlesxs/qcron"
	"github.com/charlesxs/qcron/task"
	"log"
	"time"
)

var c = flag.String("c", "", "config path")

func EchoTask(args ...interface{}) error  {
	log.Println("\033[32mI'm working hard\033[0m")
	return nil
}

func main()  {
	flag.Parse()

	// task
	t, err := task.NewTask(
		"*/2 * * * * *",
		"MyEchoTask",
		EchoTask,
		make([]interface{}, 0),
		"Task Demo",
		)
	if err != nil {
		log.Fatal(err)
	}

	err = task.Manager.Register(t)
	if err != nil {
		log.Fatal(err)
	}
	qcron.Run(*c)

	// your main program
	for {
		time.Sleep(time.Minute * 1)
	}
}

$ cat c1.toml
[cron]
listen = "127.0.0.1:8008"
nodes = ["127.0.0.1:8008", "127.0.0.1:8009", "127.0.0.1:8010"]

$ cat c2.toml
[cron]
listen = "127.0.0.1:8009"
nodes = ["127.0.0.1:8008", "127.0.0.1:8009", "127.0.0.1:8010"]

$ cat c3.toml
[cron]
listen = "127.0.0.1:8010"
nodes = ["127.0.0.1:8008", "127.0.0.1:8009", "127.0.0.1:8010"]


$  go run demo.go -c /path/to/c1.toml
$  go run demo.go -c /path/to/c2.toml
$  go run demo.go -c /path/to/c3.toml
```
