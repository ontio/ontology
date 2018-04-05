package netserver

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ontology/common/config"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

//NetServer represent all the actions in net layer
type NetServer struct {
	Self        *peer.Peer
	ReceiveChan chan types.MsgPayload
	reconnectAddrs
}

type reconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//InitListen start listening on the config port and keep on line
func (n *NetServer) Start() {
	n.Self.StartListen()
	n.connectSeeds()
	go n.keepConnection()
}

func (n *NetServer) GetVersion() uint32 {
	return n.Self.GetVersion()
}
func (n *NetServer) GetPort() uint16 {
	return n.Self.GetPort()
}
func (n *NetServer) GetConsensusPort() uint16 {
	return n.Self.GetConsensusPort()
}

//
func (n *NetServer) GetId() uint64 {
	return n.Self.GetID()
}

//GetTime return the last contact time of self peer
func (n *NetServer) GetTime() int64 {
	return n.Self.GetTime()
}

//GetState return the self peer`s state
func (n *NetServer) GetState() uint32 {
	return n.Self.GetState()
}

//GetServices return the service state of self peer
func (n *NetServer) GetServices() uint64 {
	return n.Self.GetServices()
}

//GetNeighborAddrs return all the nbr peer`s addr
func (n *NetServer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return n.Self.Np.GetNeighborAddrs()
}

//GetConnectionCnt return the total number of valid connections
func (n *NetServer) GetConnectionCnt() uint32 {
	return n.Self.Np.GetNbrNodeCnt()
}

func (n *NetServer) GetMsgCh() chan types.MsgPayload {
	return n.ReceiveChan
}

//Tx send data buf to specil peer
func (n *NetServer) Tx(id uint64, data []byte) {
	node, ok := n.Self.Np.List[id]
	if ok == true {
		node.Send(data)
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (n *NetServer) reachMinConnection() bool {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFTMINNODENUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLOMINNODENUM
	}
	return int(n.GetConnectionCnt())+1 >= minCount
}

//DisconnectNotify called when disconnect event trigger
func DisconnectNotify(v interface{}) {
	if p, ok := v.(*peer.Peer); ok {
		p.Close()
	}
}

//IsPeerEstablished return the establise state of given peer`s id
func (n *NetServer) IsPeerEstablished(id uint64) bool {
	return n.Self.Np.NodeEstablished(id)
}

//reqNbrList ask the peer for its neighbor list
func (n *NetServer) reqNbrList(*peer.Peer) {

}

//Connect begin the connect thread to given adderss
func (n *NetServer) Connect(addr string) {

}

//Halt stop all net layer logic
func (n *NetServer) Halt() {

}

//connectSeeds connect the seeds in seedlist and call for nbr list
func (n *NetServer) connectSeeds() {
	if n.reachMinConnection() {
		return
	}
	seedNodes := config.Parameters.SeedList
	for _, nodeAddr := range seedNodes {
		found := false
		var p *peer.Peer
		var ip net.IP
		n.Self.Np.Lock()
		for _, tn := range n.Self.Np.List {
			ipAddr, _ := tn.GetAddr16()
			ip = ipAddr[:]
			addrstring := ip.To16().String() + ":" + strconv.Itoa(int(tn.GetPort()))
			if nodeAddr == addrstring {
				p = tn
				found = true
				break
			}
		}
		n.Self.Np.Unlock()
		if found {
			if p.GetState() == types.ESTABLISH {
				n.reqNbrList(p)
			}
		} else { //not found
			go n.Connect(nodeAddr)
		}
	}
}

//retryInactivePeer try to connect peer in INACTIVITY state
func (n *NetServer) retryInactivePeer() {

}

//keepConnection
func (n *NetServer) keepConnection() {
	t := time.NewTimer(time.Second * types.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			n.connectSeeds()
			n.retryInactivePeer()
			t.Stop()
			t.Reset(time.Second * types.CONN_MONITOR)
		}
	}
}
