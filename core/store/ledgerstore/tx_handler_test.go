package ledgerstore

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestSyncMapRange(t *testing.T) {
	m := sync.Map{}

	for i := 0; i < 10; i++ {
		m.Store("k"+strconv.Itoa(i), "v"+strconv.Itoa(i))
	}
	cnt := 0

	m.Range(func(key, value interface{}) bool {
		fmt.Printf("key :%s, val :%s\n", key, value)
		cnt += 1
		if key == "k5" {
			return false
		}
		return true
	})

}

func TestSyncMapRW(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1000)
	m := &sync.Map{}

	m.Store("key", 1)

	for i := 0; i < 1000; i++ {
		go writeAndRead(t, m)
	}

}
func writeAndRead(t *testing.T, m *sync.Map) {
	v, _ := m.Load("key")
	m.Store("key", v.(int)+1)

}
