package node

import (
	. "GoOnchain/net/protocol"
	"fmt"
	"sync"
)

type nodeMap struct {
	Lock sync.RWMutex
	List map[string]*node
}

func (nm *nodeMap) Broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range nm.List {
		if node.state == ESTABLISH && node.relay == true {
			go node.Tx(buf)
		}
	}
}

func (nm nodeMap) containNode(node node) bool {
       _, ok := nm.List[node.id]
       return ok
}

func (nm *nodeMap) add(node *node) {
	//TODO lock the node Map
	// TODO multi client from the same IP address issue
       if (nm.containNode(*node)) {
               fmt.Printf("Insert a existed node\n")
       } else {
               nm.List[node.id] = node
       }
}

func (nm *nodeMap) delNode(node *node) {
	//TODO lock the node Map
       if (nm.containNode(*node)) {
               delete(nm.List, node.id)
       } else {
               fmt.Printf("Delete unexisted node\n")
       }
}

func (nm *nodeMap) getConnection() uint {
	//TODO lock the node Map
	var cnt uint
	for _, node := range nm.List {
		if node.state == ESTABLISH {
			cnt++
		}
	}
	return cnt
}

func (nm *nodeMap) init() {
	nm.List = make(map[string]*node)
}

func (node node) GetConnectionCnt() uint {
	return node.neighb.getConnection()
}
