package node

import (
	"DNA/common/log"
	"DNA/core/ledger"
	. "DNA/net/message"
	. "DNA/net/protocol"
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
		log.Debug("local header height is ", ledger.DefaultLedger.Store.GetHeaderHeight(), " ,other's height is ", n.GetHeight())
		if uint64(ledger.DefaultLedger.Store.GetHeaderHeight()) < n.GetHeight() {
			if n.LocalNode().IsSyncFailed() == false {
				SendMsgSyncHeaders(n)
				n.StartRetryTimer()
				break
			}
		}
	}
}

func (node node) ReqNeighborList() {
	buf, _ := NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

// Fixme the Nodes should be a parameter
func (node node) updateNodeInfo() {
	ticker := time.NewTicker(time.Second * PERIODUPDATETIME)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			//GetHeaders process haven't finished yet. So comment it now.
			node.GetBlkHdrs()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}
