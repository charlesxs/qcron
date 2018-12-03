package ndcenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"qcron/config"
	"qcron/libs/hash"
	"sync"
	"time"
)

type NDCenter struct {
	CronConfig *config.CronConfig
	Ch *hash.ConsistentHash
	currentNode map[string]int
}

var VotesContext = Votes{
	VM: make(map[string]*Vote),
}

type Votes struct {
	sync.Mutex
	VM map[string]*Vote
}

type Vote struct {
	Bill int
	Node string
	CurrentTime int64
}


func (ndc *NDCenter) NextNode(key string) (string, error) {
	var index int
	var length = len(ndc.CronConfig.Cron.Nodes)
	nodes, err := ndc.Ch.GetNodes(key, length)
	if err != nil {
		return "", err
	}

	if ndc.currentNode == nil {
		ndc.currentNode = make(map[string]int)
	}

	if _, ok := ndc.currentNode[key]; !ok {
		ndc.currentNode[key] = index
		return nodes[index], nil
	}

	index = (ndc.currentNode[key] + 1) % length
	ndc.currentNode[key] = index
	return nodes[index], nil
}

func (ndc *NDCenter) Ensure(key string, count int) bool {
	// 清空上次选举结果
	if count == 0 {
		VotesContext.VM[key] = new(Vote)
	}

	if count >= len(ndc.CronConfig.Cron.Nodes) {
		return false
	}

	node, err := ndc.NextNode(key)
	if err != nil {
		return false
	}

	if node != ndc.CronConfig.Cron.Listen {
		// ping 对方节点
		client := http.Client{Timeout: time.Millisecond * 100}
		resp, err := client.Get(node + "/ping")
		if err != nil || resp.StatusCode != 200 {
			log.Println(fmt.Sprintf("ndcener::EnsureNode node %s defunct\n", node))

			// 如果对方非存活状态, 则重新hash节点并选举
			return ndc.Ensure(key, count + 1)
		}
		return false
	}

	// 发起选举
	m := make(map[string]Vote)
	m[key] = Vote{Bill: 0, Node: node, CurrentTime: time.Now().Unix()}
	payload, _ := json.Marshal(m)

	client := http.Client{Timeout: time.Millisecond * 100}
	for _, v := range ndc.CronConfig.Cron.Nodes {
		resp, err := client.Post(v + "/vote","application/json",
			bytes.NewReader(payload))
		if err != nil || resp.StatusCode != 200 {
			log.Println(fmt.Sprintf(
				"ndcenter::Ensure send vote result failed: %s\n", err))
			continue
		}

		data, err := simplejson.NewFromReader(resp.Body)
		if err != nil {
			log.Println(fmt.Sprintf(
				"ndcenter::Ensure read response body error: %s %s\n", node, err))
			continue
		}
		_ := resp.Body.Close()

		vInterface := data.Get("data").Get(key).Interface()
		if v, ok := vInterface.(Vote); ok {
			VotesContext.VM[key] = &Vote{v.Bill, v.Node, v.CurrentTime}
		}
	}

	// 自投1票
	VotesContext.Lock()
	defer VotesContext.Unlock()
	VotesContext.VM[key].Bill++
	if VotesContext.VM[key].Bill > len(ndc.CronConfig.Cron.Nodes) / 2 {
		// 选举成功，返回true 执行
		return true
	}

	// 选举失败重新hash 并选举
	return ndc.Ensure(key, count + 1)
}

func (ndc *NDCenter) ServerRun() {
	http.HandleFunc("/ping", HandlePing)
	http.HandleFunc("/vote", HandleVote)

	if ndc.CronConfig.Cron.Listen == "" {
		log.Println("ndcenter::Run empty address, check config file please")
		return
	}

	err := http.ListenAndServe(ndc.CronConfig.Cron.Listen, nil)
	if err != nil {
		log.Println(fmt.Sprintf("ndcenter::Run http listen error: %s", err))
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
		m[k] = Vote{Bill: 0, Node: v.Node, CurrentTime: time.Now().Unix()}
		payload, _ := json.Marshal(m)

		if time.Now().Unix() - VotesContext.VM[k].CurrentTime <= 2 {
			Jsonify(w, "", 200, payload)
			return
		}

		VotesContext.VM[k].Bill = v.Bill + 1
		VotesContext.VM[k].Node = v.Node
		VotesContext.VM[k].CurrentTime = v.CurrentTime

		// 收到后将票数+1 并返回给发起方
		m[k] = Vote{Bill: 0, Node: v.Node, CurrentTime: time.Now().Unix()}
		payload, _ = json.Marshal(m)
		Jsonify(w, "", 200, payload)
	}
}

func Jsonify(w http.ResponseWriter, msg string, code int, data interface{}) {
	resp := make(map[string]interface{})
	resp["code"] = code
	resp["message"] = msg
	resp["data"] = data
	rdata, err := json.Marshal(resp)
	if err != nil {
		log.Println(fmt.Sprintf("ndcenter::Jsonify error: %d %s %v", code, msg, data))
		return
	}
	_, err = fmt.Fprint(w, string(rdata))
	if err != nil {
		log.Println(fmt.Sprintf("ndcenter::Jsonify write error: %s", err))
	}
}



