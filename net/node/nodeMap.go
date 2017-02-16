package node

import (
	. "GoOnchain/net/protocol"
	"math/rand"
	"sync"
	"time"
)

type nodeMap struct {
	Lock sync.RWMutex
	List map[string]*node
}

func (Nodes *nodeMap) Broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range Nodes.List {
		if node.state == ESTABLISH && node.relay == true {
			go node.Tx(buf)
		}
	}
}

func (Nodes *nodeMap) add(node *node) {
	//TODO lock the node Map
	// TODO check whether the node existed or not
	// TODO dupicate IP address Nodes issue
	Nodes.List[node.id] = node
	// Unlock the map
}

func (Nodes *nodeMap) delNode(node *node) {
	//TODO lock the node Map
	delete(Nodes.List, node.id)
	// Unlock the map
}

func (Nodes *nodeMap) getConnection() uint {
	//TODO lock the node Map
	var cnt uint
	for _, node := range Nodes.List {
		if node.state == ESTABLISH {
			cnt++
		}
	}
	return cnt
}

func (Nodes *nodeMap) init() {
	Nodes.List = make(map[string]*node)
}

func (node node) GetConnectionCnt() uint {
	return node.neighb.getConnection()
}

func InitNodes() {
	// TODO write lock
	n := NewNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.port = NODETESTPORT
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	// Fixme replace with the real random number
	n.nonce = rand.Uint32()
}
