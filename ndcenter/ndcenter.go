package ndcenter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/charlesxs/qcron/config"
	"github.com/charlesxs/qcron/libs"
	"github.com/charlesxs/qcron/libs/hash"
	"github.com/charlesxs/qcron/task"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type (
	NDCenter struct {
		CronConfig *config.CronConfig
		Ch *hash.ConsistentHash
	}

	hashCache struct {
		nodes []string
		cursor int
	}

	Vote struct {
		Bill int
		Tof int 		// 第几轮投票
		CurrentTime int64
	}
)

var HashCache = struct {
	sync.Mutex
	m map[string]*hashCache
}{m: make(map[string]*hashCache)}

var VotesContext = struct {
	sync.Mutex
	VM map[string]*Vote
}{VM: make(map[string]*Vote)}


func (ndc *NDCenter) Init() {
	log.Println("ndcenter::Init start init cluster")
	data := map[string]interface{}{
		"node": ndc.CronConfig.Cron.Listen,
	}

	tasks := make(map[string]time.Time)
	for _, t := range task.Manager.Tasks {
		tasks[t.TaskID] = t.TaskTime.NextExecTime
	}

	// write self
	libs.InfoCache.WriteCache(tasks)
	data["tasks"] = tasks
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("ndcenter::Init init failed, error: %s\n", err)
		return
	}
	client := http.Client{}
	for _, node := range ndc.CronConfig.Cron.Nodes {
		if node == ndc.CronConfig.Cron.Listen {
			continue
		}

		url := fmt.Sprintf("http://%s/sync", node)
		resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
		if err != nil || resp.StatusCode != 200 {
			log.Printf("ndcenter::Init %s", err)
			continue
		}

		m := make(map[string]interface{})
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ndcenter::Init response error from node: %s, error: %s, respData: %v\n",
				node, err, respData)
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			log.Printf("ndcenter::Init close response body error: %s\n", err)
			continue
		}

		err = json.Unmarshal(respData, &m)
		if err != nil {
			log.Printf("ndcenter::Init bad response body, err: %s, respData: %v\n", err, respData)
			continue
		}

		taskTimes := make(map[string]time.Time)
		if d, ok := m["data"].(map[string]interface{}); ok {
			for k, v := range d {
				if t, err := parseTime(v); err != nil {
					log.Printf("ndcenter::Init bad response body, error: %s", err)
				} else {
					taskTimes[k] = t
				}
			}

			libs.InfoCache.WriteCache(taskTimes)
		} else {
			log.Printf("ndcenter::Init parse response error, data: %v", m)
		}
	}

	var ok bool
	for i := 0; i < 60; i++ {
		libs.InfoCache.ForEach(func(k string, v []time.Time) bool {
			if v != nil && len(v) >= len(ndc.CronConfig.Cron.Nodes){
				ok = true
				return false
			}
			return true
		})
		if ok {
			break
		}
		time.Sleep(time.Second * 1)
	}

	task.UpdateTasks()
	if ok {
		log.Println("ndcenter::Init init cluster ok")
		return
	}

	log.Println("ndcenter::Init init cluster failed")
}

func (ndc *NDCenter) GetNode(key string, next bool) (string, error) {
	var index int

	HashCache.Lock()
	defer HashCache.Unlock()
	if _, ok := HashCache.m[key]; !ok {
		nodes, err := ndc.Ch.GetNodes(key, len(ndc.CronConfig.Cron.Nodes))
		if err != nil {
			return "", err
		}

		HashCache.m[key] = &hashCache{
			nodes: nodes,
			cursor: index,
		}
	}

	if next {
		index = (HashCache.m[key].cursor + 1) % len(HashCache.m[key].nodes)
		HashCache.m[key].cursor = index
	}

	return HashCache.m[key].nodes[HashCache.m[key].cursor], nil

}

func (ndc *NDCenter) Ensure(key string, count int) (bool, int) {
	var (
		node string
		err error
		next bool
	)
	// 清空上次选举结果
	if count == 0 {
		VotesContext.VM[key] = new(Vote)
	}
	if count >= len(ndc.CronConfig.Cron.Nodes) {
		return false, count
	}

	if count > 0 {
		next = true
	}

	node, err = ndc.GetNode(key, next)
	if err != nil {
		return false, count
	}

	if node != ndc.CronConfig.Cron.Listen {
		// ping 对方节点
		client := http.Client{Timeout: time.Millisecond * 200}
		url := fmt.Sprintf("http://%s/ping", node)
		resp, err := client.Get(url)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("ndcenter::Ensure node %s defunct\n", node)

			// 如果对方非存活状态, 则重新hash节点并选举
			return ndc.Ensure(key, count + 1)
		}
		return false, count
	}

	// 发起选举
	m := make(map[string]Vote)
	m[key] = Vote{Bill: 0, Tof: count, CurrentTime: time.Now().Unix()}
	payload, _ := json.Marshal(m)

	client := http.Client{}
	for _, v := range ndc.CronConfig.Cron.Nodes {
		if v == ndc.CronConfig.Cron.Listen {
			continue
		}

		url := fmt.Sprintf("http://%s/vote", v)
		resp, err := client.Post(url,"application/json",
			bytes.NewReader(payload))
		if err != nil || resp.StatusCode != 200 {
			log.Printf(
				"ndcenter::Ensure send vote result failed: %s\n", err)
			continue
		}

		data, err := simplejson.NewFromReader(resp.Body)
		if err != nil {
			log.Printf(
				"ndcenter::Ensure read response body error: %s %s\n", node, err)
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			log.Printf("ndcenter::Ensure close reponse body error: %s\n", err)
		}

		bill, err := data.Get("data").Get(key).Get("Bill").Int()
		if err != nil {
			log.Printf("ndcenter::Ensure response body error: %s, body: %v", err, data)
		}

		VotesContext.Lock()
		VotesContext.VM[key].Bill += bill
		VotesContext.VM[key].Tof = count
		VotesContext.VM[key].CurrentTime = time.Now().Unix()
		VotesContext.Unlock()
	}

	// 自投1票
	VotesContext.Lock()
	defer VotesContext.Unlock()
	VotesContext.VM[key].Bill++
	if VotesContext.VM[key].Bill > len(ndc.CronConfig.Cron.Nodes) / 2 {
		// 选举成功，返回true 执行
		return true, count
	}

	// 选举失败重新hash 并选举
	return ndc.Ensure(key, count + 1)
}

func (ndc *NDCenter) ServerRun() {
	http.HandleFunc("/ping", HandlePing)
	http.HandleFunc("/vote", HandleVote)
	http.HandleFunc("/sync", HandleSyncInfo)

	if ndc.CronConfig.Cron.Listen == "" {
		log.Println("ndcenter::Run empty address, check config file please")
		return
	}

	err := http.ListenAndServe(ndc.CronConfig.Cron.Listen, nil)
	if err != nil {
		log.Printf("ndcenter::Run http listen error: %s\n", err)
		return
	}
}

func HandlePing(w http.ResponseWriter, req *http.Request)  {
	Jsonify(w, "pong", 200, nil)
}


func HandleVote(w http.ResponseWriter, req *http.Request) {
	var voteData = make(map[string]Vote)
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		Jsonify(w, fmt.Sprintf("ndcenter::HandleVote read body error: %s", err), 500, nil)
		return
	}

	defer func() {
		err := req.Body.Close()
		if err != nil {
			log.Printf("ndcenter::HandleVote close request body error: %s\n", err)
		}
	}()

	err = json.Unmarshal(data, &voteData)
	if err != nil {
		msg := fmt.Sprintf("ndcenter::HandleVote request body error: %s", err)
		Jsonify(w, msg, 500, nil)
		log.Println(msg)
		return
	}

	// 选举
	VotesContext.Lock()
	defer VotesContext.Unlock()
	for k, v := range voteData {
		if _, ok := VotesContext.VM[k]; !ok {
			VotesContext.VM[k] = new(Vote)
		}

		// result
		m := make(map[string]Vote)
		m[k] = Vote{Bill: 0, Tof: 0, CurrentTime: time.Now().Unix()}

		// 第1轮投票且时间间隔小于2s, 主要防止脑裂情况下, 只响应最先收到的投票请求
		if v.Tof == 0 && time.Now().Unix() - VotesContext.VM[k].CurrentTime <= 2 {
			Jsonify(w, "", 200, m)
			return
		}

		VotesContext.VM[k].Bill = v.Bill + 1
		VotesContext.VM[k].Tof = v.Tof
		VotesContext.VM[k].CurrentTime = v.CurrentTime

		// 收到后将票数+1 并返回给发起方
		m[k] = Vote{Bill: v.Bill + 1, Tof: v.Tof, CurrentTime: time.Now().Unix()}
		Jsonify(w, "", 200, m)
	}
}

func HandleSyncInfo(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		Jsonify(w, fmt.Sprintf("%s", err), 500, nil)
		return
	}

	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Printf("ndcenter::HandleSyncInfo close request body error: %s\n", err)
		}
	}()

	postData := make(map[string]interface{})
	err = json.Unmarshal(data, &postData)
	if err != nil {
		msg := fmt.Sprintf("ndcenter::HandleSyncInfo request body error: %s\n", err)
		Jsonify(w, msg, 500, nil)
		log.Println(msg)
		return
	}

	node := postData["node"].(string)
	tasks := make(map[string]time.Time)
	if d, ok := postData["tasks"].(map[string]interface{}); ok {
		for k, v := range d {
			if t, err := parseTime(v); err != nil {
				log.Printf("ndcenter::HandleSyncInfo bad request body: %v, error: %s", postData, err)
			} else {
				tasks[k] = t
			}
		}
	}
	libs.InfoCache.WriteCache(tasks)
	task.UpdateTasks()

	// 当node 重新上线后复位所有此node上的任务
	HashCache.Lock()
	for _, v := range HashCache.m {
		for i := 0; i < v.cursor; i++ {
			if v.nodes[i] == node {
				v.cursor = i
				break
			}
		}
	}
	HashCache.Unlock()

	m := make(map[string]time.Time)
	for _, t := range task.Manager.Tasks {
		m[t.TaskID] = t.TaskTime.NextExecTime
	}

	Jsonify(w, "success", 200, m)
}

func Jsonify(w http.ResponseWriter, msg string, statusCode int, data interface{}) {
	resp := make(map[string]interface{})
	resp["code"] = statusCode
	resp["message"] = msg
	resp["data"] = data
	rdata, err := json.Marshal(resp)
	if err != nil {
		log.Printf("ndcenter::Jsonify error: %d %s %v\n", statusCode, msg, data)
		return
	}

	w.WriteHeader(statusCode)
	_, err = fmt.Fprint(w, string(rdata))
	if err != nil {
		log.Printf("ndcenter::Jsonify write error: %s\n", err)
	}
}

func parseTime(v interface{}) (time.Time, error) {
	if t, ok := v.(string); ok {
		return time.Parse(time.RFC3339Nano, t)
	}
	return time.Time{}, errors.New("expect time string")
}

