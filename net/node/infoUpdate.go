package node

import (
	"DNA/common/config"
	"DNA/common/log"
	"DNA/core/ledger"
	. "DNA/net/message"
	. "DNA/net/protocol"
	"fmt"
	"net"
	"strconv"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node *node) GetBlkHdrs() {
	if node.local.GetNbrNodeCnt() < MINCONNCNT {
		return
	}

	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if uint64(ledger.DefaultLedger.Store.GetHeaderHeight()) < n.GetHeight() {
			if n.LocalNode().IsSyncFailed() == false {
				SendMsgSyncHeaders(n)
				n.StartRetryTimer()
				break
			}
		}
	}
}

func (node *node) SyncBlk() {
	headerHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	currentBlkHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	if currentBlkHeight >= headerHeight {
		return
	}
	var dValue int32
	var reqCnt uint32
	var i uint32
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		n.RemoveFlightHeightLessThan(currentBlkHeight)
		count := MAXREQBLKONCE - uint32(n.GetFlightHeightCnt())
		dValue = int32(headerHeight - currentBlkHeight - reqCnt)
		for i = 1; i <= count && dValue >= 0; i++ {
			hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(currentBlkHeight + reqCnt)
			ReqBlkData(n, hash)
			n.StoreFlightHeight(currentBlkHeight + reqCnt)
			reqCnt++
			dValue--
		}
	}
}

func (node *node) SendPingToNbr() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		t := n.GetLastRXTime()
		if time.Since(t).Seconds() > PERIODUPDATETIME {
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
}

func (node *node) HeartBeatMonitor() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			t := n.GetLastRXTime()
			if time.Since(t).Seconds() > (PERIODUPDATETIME * KEEPALIVETIMEOUT) {
				log.Warn("keepalive timeout!!!")
				n.SetState(INACTIVITY)
				//n.CloseConn()
			}
		}
	}
}

func (node *node) ReqNeighborList() {
	buf, _ := NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

func (node *node) ConnectSeeds() {
	if node.nbrNodes.GetConnectionCnt() == 0 {
		seedNodes := config.Parameters.SeedList
		for _, nodeAddr := range seedNodes {
			go node.Connect(nodeAddr)
		}
	}
}

func getNodeAddr(n *node) NodeAddr {
	var addr NodeAddr
	addr.IpAddr, _ = n.GetAddr16()
	addr.Time = n.GetTime()
	addr.Services = n.Services()
	addr.Port = n.GetPort()
	addr.ID = n.GetID()
	return addr
}

func (node *node) reconnect(peer *node) error {
	isTls := config.Parameters.IsTLS
	addr := getNodeAddr(peer)
	var ip net.IP
	ip = addr.IpAddr[:]
	nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
	log.Info("try to reconnect peer, peer addr is ", nodeAddr)
	var conn net.Conn
	var err error
	if isTls {
		conn, err = TLSDial(nodeAddr)
		if err != nil {
			return nil
		}
	} else {
		conn, err = NonTLSDial(nodeAddr)
		if err != nil {
			return nil
		}
	}
	peer.tryTimes = 0
	peer.conn = conn
	peer.addr, err = parseIPaddr(conn.RemoteAddr().String())
	peer.local = node

	log.Info(fmt.Sprintf("Reconnect node %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))
	go peer.rx()

	peer.SetState(ESTABLISH)

	return nil
}

func (node *node) TryConnect() {
	for _, n := range node.nbrNodes.List {
		if n.GetState() == INACTIVITY && n.tryTimes < 3 {
			//try to connect
			n.tryTimes++
			node.reconnect(n)
		}
	}

}

// FIXME part of node info update function could be a node method itself intead of
// a node map method
// Fixme the Nodes should be a parameter
func (node *node) updateNodeInfo() {
	ticker := time.NewTicker(time.Second * PERIODUPDATETIME)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			node.ConnectSeeds()
			node.SendPingToNbr()
			node.GetBlkHdrs()
			node.SyncBlk()
			node.HeartBeatMonitor()
			node.TryConnect()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}
