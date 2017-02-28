package node

import (
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node node) GetBlkHdrs() {
	for _, n := range node.local.neighb.List {
		h1 := n.GetHeight()
		h2:= node.local.GetLedger().GetLocalBlockChainHeight()
		//fmt.Printf("The neighbor height is %d, local height is %d\n", h1, h2)
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
			//common.Trace()
			node.GetBlkHdrs()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}
