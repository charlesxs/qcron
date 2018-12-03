package ndcenter

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNDCenter_Run(t *testing.T) {
	//m := make(map[string]interface{})
	//n := make([]string, 3)
	//n[0] = "xxx"
	//m["name"] = "xs.xiao"
	//m["age"] = 12
	//m["data"] = n
	//d, _ := json.Marshal(m)
	//fmt.Println(string(d))

	//v := Vote{Bill: 1}
	//d, _ := json.Marshal(v)
	//fmt.Println(string(d))

	m := make(map[string]Vote)
	m["name"] = Vote{Bill: 1, Node: "xxxx"}
	s, _ := json.Marshal(m)
	fmt.Println(string(s))


	n := make(map[string]Vote)
	err := json.Unmarshal(s, &n)
	fmt.Println(n["name"], err)
}

