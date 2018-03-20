package node

import (
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"math/rand"
	"net"
	//"github.com/Ontology/core/ledger"
	"github.com/Ontology/p2pserver/actor"
	. "github.com/Ontology/p2pserver/message"
	. "github.com/Ontology/p2pserver/protocol"
	"strconv"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node *node) GetBlkHdrs() {
	//TODO
	if !node.IsUptoMinNodeCount() {
		return
	}
	noders := node.local.GetNeighborNoder()
	if len(noders) == 0 {
		return
	}
	nodelist := []Noder{}
	for _, v := range noders {
		height, _ := actor.GetCurrentHeaderHeight()
		if uint64(height) < v.GetHeight() {
			nodelist = append(nodelist, v)
		}
	}
	ncout := len(nodelist)
	if ncout == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(ncout)
	n := nodelist[index]
	SendMsgSyncHeaders(n)
}

func (node *node) SyncBlk() {
	//headerHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	headerHeight, _ := actor.GetCurrentHeaderHeight()
	//currentBlkHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	currentBlkHeight, _ := actor.GetCurrentBlockHeight()
	if currentBlkHeight >= headerHeight {
		return
	}
	var dValue int32
	var reqCnt uint32
	var i uint32
	noders := node.local.GetNeighborNoder()

	for _, n := range noders {
		if uint32(n.GetHeight()) <= currentBlkHeight {
			continue
		}
		n.RemoveFlightHeightLessThan(currentBlkHeight)
		count := MAXREQBLKONCE - uint32(n.GetFlightHeightCnt())
		dValue = int32(headerHeight - currentBlkHeight - reqCnt)
		flights := n.GetFlightHeights()
		if count == 0 {
			for _, f := range flights {
				//hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(f)
				hash, _ := actor.GetBlockHashByHeight(f)
				isContainBlock, _ := actor.IsContainBlock(hash)
				if isContainBlock == false {
					ReqBlkData(n, hash)
				}
			}

		}
		for i = 1; i <= count && dValue >= 0; i++ {
			//hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(currentBlkHeight + reqCnt)
			hash, _ := actor.GetBlockHashByHeight(currentBlkHeight + reqCnt)
			isContainBlock, _ := actor.IsContainBlock(hash)
			if isContainBlock == false {
				ReqBlkData(n, hash)
				n.StoreFlightHeight(currentBlkHeight + reqCnt)
			}
			reqCnt++
			dValue--
		}
	}
}

func (node *node) SendPingToNbr() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			buf, err := NewPingMsg()
			if err != nil {
				log.Error("failed build a new ping message")
			} else {
				go n.Tx(buf)
			}
		}
	}
}

func (node *node) HeartBeatMonitor() {
	noders := node.local.GetNeighborNoder()
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodUpdateTime = config.Parameters.GenBlockTime / TIMESOFUPDATETIME
	} else {
		periodUpdateTime = config.DEFAULTGENBLOCKTIME / TIMESOFUPDATETIME
	}
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			t := n.GetLastRXTime()
			if t.Before(time.Now().Add(-1 * time.Second * time.Duration(periodUpdateTime) * KEEPALIVETIMEOUT)) {
				log.Warn("keepalive timeout!!!")
				n.SetState(INACTIVITY)
				n.CloseConn()
			}
		}
	}
}

func (node *node) ReqNeighborList() {
	buf, _ := NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

func (node *node) ConnectSeeds() {
	if node.IsUptoMinNodeCount() {
		return
	}
	seedNodes := config.Parameters.SeedList
	for _, nodeAddr := range seedNodes {
		found := false
		var n Noder
		var ip net.IP
		node.nbrNodes.Lock()
		for _, tn := range node.nbrNodes.List {
			addr := getNodeAddr(tn)
			ip = addr.IpAddr[:]
			addrstring := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
			if nodeAddr == addrstring {
				n = tn
				found = true
				break
			}
		}
		node.nbrNodes.Unlock()
		if found {
			if n.GetState() == ESTABLISH {
				n.ReqNeighborList()
			}
		} else { //not found
			go node.Connect(nodeAddr, false)
		}
	}
}

func getNodeAddr(n *node) NodeAddr {
	var addr NodeAddr
	addr.IpAddr, _ = n.GetAddr16()
	addr.Time = n.GetTime()
	addr.Services = n.Services()
	addr.Port = n.GetPort()
	addr.ConsensusPort = n.GetConsensusPort()
	addr.ID = n.GetID()
	return addr
}

func (node *node) reconnect() {
	node.RetryConnAddrs.Lock()
	defer node.RetryConnAddrs.Unlock()
	lst := make(map[string]int)
	for addr := range node.RetryAddrs {
		node.RetryAddrs[addr] = node.RetryAddrs[addr] + 1
		rand.Seed(time.Now().UnixNano())
		log.Trace("Try to reconnect peer, peer addr is ", addr)
		<-time.After(time.Duration(rand.Intn(CONNMAXBACK)) * time.Millisecond)
		log.Trace("Back off time`s up, start connect node")
		node.Connect(addr, false)
		if node.RetryAddrs[addr] < MAXRETRYCOUNT {
			lst[addr] = node.RetryAddrs[addr]
		}
	}
	node.RetryAddrs = lst

}

func (n *node) TryConnect() {
	if n.fetchRetryNodeFromNeiborList() > 0 {
		n.reconnect()
	}
}

func (n *node) fetchRetryNodeFromNeiborList() int {
	n.nbrNodes.Lock()
	defer n.nbrNodes.Unlock()
	var ip net.IP
	neibornodes := make(map[uint64]*node)
	for _, tn := range n.nbrNodes.List {
		addr := getNodeAddr(tn)
		ip = addr.IpAddr[:]
		nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
		if tn.GetState() == INACTIVITY {
			//add addr to retry list
			n.AddInRetryList(nodeAddr)
			//close legacy node
			if tn.conn != nil {
				tn.CloseConn()
			}
		} else {
			//add others to tmp node map
			n.RemoveFromRetryList(nodeAddr)
			neibornodes[tn.GetID()] = tn
		}
	}
	n.nbrNodes.List = neibornodes
	return len(n.RetryAddrs)
}

// FIXME part of node info update function could be a node method itself intead of
// a node map method
// Fixme the Nodes should be a parameter
func (node *node) updateNodeInfo() {
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodUpdateTime = config.Parameters.GenBlockTime / TIMESOFUPDATETIME
	} else {
		periodUpdateTime = config.DEFAULTGENBLOCKTIME / TIMESOFUPDATETIME
	}
	ticker := time.NewTicker(time.Second * (time.Duration(periodUpdateTime)))
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			node.SendPingToNbr()
			node.GetBlkHdrs()
			node.SyncBlk()
			node.HeartBeatMonitor()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}

func (node *node) updateConnection() {
	t := time.NewTimer(time.Second * CONNMONITOR)
	for {
		select {
		case <-t.C:
			node.ConnectSeeds()
			node.TryConnect()
			t.Stop()
			t.Reset(time.Second * CONNMONITOR)
		}
	}

}
