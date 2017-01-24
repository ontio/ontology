package node

import (
	"sync"
	"time"
	"math/rand"
	. "GoOnchain/net/protocol"
)

type nodeMap struct {
	Node *node
	Lock sync.RWMutex
	List map[string]*node
}

// FIXME node table should be dynamic create instead of static define here
// TODO: All method relative with the node map should be in a seperated file
var Nodes nodeMap

func (Nodes *nodeMap) broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range Nodes.List {
		if node.state == ESTABLISH {
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

func InitNodes() {
	// TODO write lock
	n := newNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.port = NODETESTPORT
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	n.nonce = rand.Uint32()

	Nodes.Node = n
	Nodes.List = make(map[string]*node)
}
