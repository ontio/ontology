package node

import (
	"DNA/core/ledger"
	. "DNA/net/message"
	. "DNA/net/protocol"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node *node) GetBlkHdrs() {
	node.local.nbrNodes.RLock()
	defer node.local.nbrNodes.RUnlock()
	for _, n := range node.local.List {
		h1 := n.GetHeight()
		h2 := ledger.DefaultLedger.GetLocalBlockChainHeight()
		if (node.GetState() == ESTABLISH) && (h1 > uint64(h2)) {
			buf, _ := NewMsg("getheaders", node.local)
			go node.Tx(buf)
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
			//log.Trace()
			node.GetBlkHdrs()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}
