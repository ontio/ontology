package node

import (
	. "GoOnchain/net/protocol"
	"fmt"
	"sync"
)

// The neigbor node list
type nbrNodes struct {
	Lock sync.RWMutex
	List map[uint64]*node
}

func (nm *nbrNodes) Broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range nm.List {
		if node.state == ESTABLISH && node.relay == true {
			go node.Tx(buf)
		}
	}
}

func (nm *nbrNodes) NodeExisted(uid uint64) bool {
	_, ok := nm.List[uid]
	return ok
}

func (nm *nbrNodes) AddNbrNode(n Noder) {
	//TODO lock the node Map
	// TODO multi client from the same IP address issue
	if (nm.NodeExisted(n.GetID())) {
               fmt.Printf("Insert a existed node\n")
	} else {
		node, err := n.(*node)
		if (err == false) {
			fmt.Println("Convert the noder error when add node")
			return
		}
		nm.List[n.GetID()] = node
	}
}

func (nm *nbrNodes) DelNbrNode(id uint64) (Noder, bool) {
	//TODO lock the node Map
	n, ok := nm.List[id]
	if (ok == false) {
		return nil, false
	}
	delete(nm.List, id)
	return n, true
}

func (nm nbrNodes) GetConnectionCnt() uint {
	//TODO lock the node Map
	var cnt uint
	for _, node := range nm.List {
		if node.state == ESTABLISH {
			cnt++
		}
	}
	return cnt
}

func (nm *nbrNodes) init() {
	nm.List = make(map[uint64]*node)
}

func (nm nbrNodes) NodeEstablished(id uint64) bool {
	n, ok := nm.List[id]
	if (ok == false) {
		return false
	}

	if (n.state != ESTABLISH) {
		return false
	}

	return true
}
