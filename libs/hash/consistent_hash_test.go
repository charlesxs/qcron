package hash

import (
	"testing"
	"fmt"
)

func TestConsistentHash(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}
	ch := NewConsistentHash(hosts, 100)

	fmt.Println(ch.GetNode("abc"))
	fmt.Println(ch.GetNodes("abc", 3))
}


func TestConsistentHash_RemoveNode(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}
	ch := NewConsistentHash(hosts, 100)

	fmt.Println(ch.RemoveNode("host3"))
	fmt.Println(ch.Nodes)
}