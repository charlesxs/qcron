package ndcenter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestNDCenter_Ensure(t *testing.T) {
	//c, err := config.NewConfig("/Users/charles/shells/go/src/qcron/config/config.toml")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//ndc := &NDCenter{
	//	CronConfig: c,
	//	Ch: hash.NewConsistentHash(c.Cron.Nodes, 100),
	//}
	//
	//go ndc.ServerRun()
	//fmt.Println(ndc.Ensure("MyTask1", 0))

	n := make(map[string]time.Time)
	data := make(map[string]time.Time)
	data["task1"] = time.Now()
	rd, _ := json.Marshal(data)
	fmt.Println(string(rd))

	err := json.Unmarshal(rd, &n)
	fmt.Println(err, n)
}