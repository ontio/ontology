package node

import (
	. "DNA/net/protocol"
	"fmt"
	"sync"
)

// The neigbor node list
type nbrNodes struct {
	sync.RWMutex
	// Todo using the Pool structure
	List map[uint64]*node
}

func (nm *nbrNodes) Broadcast(buf []byte) {
	nm.RLock()
	defer nm.RUnlock()
	for _, node := range nm.List {
		if node.state == ESTABLISH && node.relay == true {
			// The routie need lock too
			go node.Tx(buf)
		}
	}
}

func (nm *nbrNodes) NodeExisted(uid uint64) bool {
	_, ok := nm.List[uid]
	return ok
}

func (nm *nbrNodes) AddNbrNode(n Noder) {
	nm.Lock()
	defer nm.Unlock()

	if nm.NodeExisted(n.GetID()) {
		fmt.Printf("Insert a existed node\n")
	} else {
		node, err := n.(*node)
		if err == false {
			fmt.Println("Convert the noder error when add node")
			return
		}
		nm.List[n.GetID()] = node
	}
}

func (nm *nbrNodes) DelNbrNode(id uint64) (Noder, bool) {
	nm.Lock()
	defer nm.Unlock()

	n, ok := nm.List[id]
	if ok == false {
		return nil, false
	}
	delete(nm.List, id)
	return n, true
}

func (nm *nbrNodes) GetConnectionCnt() uint {
	nm.RLock()
	defer nm.RUnlock()

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

func (nm *nbrNodes) NodeEstablished(id uint64) bool {
	nm.RLock()
	defer nm.RUnlock()

	n, ok := nm.List[id]
	if ok == false {
		return false
	}

	if n.state != ESTABLISH {
		return false
	}

	return true
}

func (node *node) GetNeighborAddrs() ([]NodeAddr, uint64) {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()

	var i uint64
	var addrs []NodeAddr
	for _, n := range node.nbrNodes.List {
		if n.GetState() != ESTABLISH {
			continue
		}
		var addr NodeAddr
		addr.IpAddr, _ = n.GetAddr16()
		addr.Time = n.GetTime()
		addr.Services = n.Services()
		addr.Port = n.GetPort()
		addr.ID = n.GetID()
		addrs = append(addrs, addr)

		i++
	}

	return addrs, i
}

func (node *node) GetNeighborHeights() ([]uint64, uint64) {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()

	var i uint64
	var heights []uint64
	heights = make([]uint64, 1)
	for _, n := range node.nbrNodes.List {
		if n.GetState() == ESTABLISH {
			height := n.GetHeight()
			heights = append(heights, height)
			i++
		}
	}
	return heights, i
}
