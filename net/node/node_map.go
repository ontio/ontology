/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package node

import (
	"fmt"
	"github.com/Ontology/common/config"
	. "github.com/Ontology/net/protocol"
	"strings"
	"sync"
)

// The neighbor node list
type nbrNodes struct {
	sync.RWMutex
	List map[uint64]*node
}

func (nm *nbrNodes) Broadcast(buf []byte) {
	nm.RLock()
	defer nm.RUnlock()
	for _, node := range nm.List {
		if node.state == ESTABLISH && node.relay == true {
			node.Tx(buf)
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
	heights := []uint64{}
	for _, n := range node.nbrNodes.List {
		if n.GetState() == ESTABLISH {
			height := n.GetHeight()
			heights = append(heights, height)
			i++
		}
	}
	return heights, i
}

func (node *node) GetNeighborNoder() []Noder {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()

	nodes := []Noder{}
	for _, n := range node.nbrNodes.List {
		if n.GetState() == ESTABLISH {
			node := n
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (node *node) GetNbrNodeCnt() uint32 {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()
	var count uint32
	for _, n := range node.nbrNodes.List {
		if n.GetState() == ESTABLISH {
			count++
		}
	}
	return count
}

func (node *node) IsUptoMinNodeCount() bool {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFT_MIN_NODE_NUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLO_MIN_NODE_NUM
	}
	return int(node.GetNbrNodeCnt())+1 >= minCount
}
