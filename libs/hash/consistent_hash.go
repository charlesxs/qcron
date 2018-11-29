package hash

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-collections/collections/set"
	"qcron/libs/bisect"
	"strconv"
)

type ConsistentHash struct {
	Nodes *set.Set
	ReplicaCount int
	hashRing Ring
}

type hashEntry struct {
	node string
	position uint64
}

type Ring []hashEntry

func (r Ring) Len() int {
	return len(r)
}

func (r Ring) Less(v interface{}, i int) bool {
	value, ok := v.(hashEntry)
	if !ok {
		return false
	}

	if value.position < r[i].position {
		return true
	}
	return false
}

func (r Ring) LessEqual(v interface{}, i int) bool  {
	value, ok := v.(hashEntry)
	if !ok {
		return false
	}

	if value.position <= r[i].position {
		return true
	}
	return false
}

func (r Ring) Insert(v interface{}, i int) (interface{}, error)  {
	value, ok := v.(hashEntry)
	if !ok {
		return nil, errors.New("must be hashEntry type")
	}

	var newList = make(Ring, len(r) + 1)
	copy(newList[:i], r[:i])
	newList[i] = value
	copy(newList[i+1:], r[i:])
	return newList, nil
}


func (ch *ConsistentHash) computeRingPosition(key []byte) uint64 {
	m := md5.New()
	m.Write(key)
	hashString := hex.EncodeToString(m.Sum(nil))
	pos, _ := strconv.ParseUint(hashString[:4], 16, 64)
	return pos
}

func (ch *ConsistentHash) AddNode(node string) error {
	var (
		key string
		pos uint64
	)

	ch.Nodes.Insert(node)
	for i := 0; i < ch.ReplicaCount; i++ {
		key = fmt.Sprintf("%s::%d", node, i)
		pos = ch.computeRingPosition([]byte(key))

		entry := hashEntry{
			node: node,
			position: pos,
		}
		newRing, err := bisect.InsertRight(ch.hashRing, entry)
		if err != nil {
			return err
		}

		ch.hashRing = newRing.(Ring)
	}
	return nil
}

func (ch *ConsistentHash) RemoveNode(node string) error  {
	ch.Nodes.Remove(node)

	newRing := make(Ring, 0, len(ch.hashRing) - 100)
	for _, v := range ch.hashRing {
		if v.node != node {
			newRing = append(newRing, v)
		}
	}

	ch.hashRing = newRing
	return nil
}

func (ch *ConsistentHash) GetNodes(key string, n int) ([]string, error)  {
	var result = make([]string, 0, n)
	if len(ch.hashRing) < 1 {
		return nil, errors.New("empty ring")
	}

	nodes, ringLength := set.New(), len(ch.hashRing)
	pos := ch.computeRingPosition([]byte(key))

	searchEntry := hashEntry{node: "", position: pos}
	index := bisect.SearchInsertPostLeft(ch.hashRing, searchEntry) % ringLength
	lastIndex := (index - 1) % ringLength

	for index != lastIndex && n > 0 {
		entry := ch.hashRing[index]
		index = (index + 1) % ringLength
		if nodes.Has(entry.node) {
			continue
		}
		nodes.Insert(entry.node)
		result = append(result, entry.node)
		n--
	}

	return result, nil
}

func (ch *ConsistentHash) GetNode(key string) (string, error)  {
	r, err := ch.GetNodes(key, 1)
	if err != nil {
		return "", err
	}
	return r[0], err
}

func NewConsistentHash(nodes []string, replicaCount int) *ConsistentHash  {
	ch := &ConsistentHash{
		Nodes: set.New(),
		ReplicaCount: replicaCount,
		hashRing: make(Ring, 0),
	}

	for _, v := range nodes {
		ch.AddNode(v)
	}
	return ch
}

