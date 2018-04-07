package netserver

import (
	"errors"
	"sync"

	"github.com/Ontology/common/log"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

//NetServer represent all the actions in net layer
type NetServer struct {
	Self        *peer.Peer
	ReceiveChan chan types.MsgPayload
	reconnectAddrs
}

//reconnectAddrs contain addr need to reconnect
type reconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//InitListen start listening on the config port and keep on line
func (n *NetServer) Start() {
	n.Self.StartListen()
}

//GetVersion return self peer`s version
func (n *NetServer) GetVersion() uint32 {
	return n.Self.GetVersion()
}

//GetPort return self peer`s txn port
func (n *NetServer) GetPort() uint16 {
	return n.Self.GetPort()
}

//GetConsensusPort return self peer`s consensus port
func (n *NetServer) GetConsensusPort() uint16 {
	return n.Self.GetConsensusPort()
}

//GetId return peer`s id
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

//Tx send data buf to peer
func (n *NetServer) Send(p *peer.Peer, data []byte, isConsensus bool) error {
	if p != nil {
		return p.Send(data, isConsensus)
	}
	log.Error("send to a invalid peer")
	return errors.New("send to a invalid peer")
}

//DisconnectNotify called when disconnect event trigger
func DisconnectNotify(v interface{}) {
	if p, ok := v.(*peer.Peer); ok {
		p.Close()
	}
}

//IsPeerEstablished return the establise state of given peer`s id
func (n *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return n.Self.Np.NodeEstablished(p.GetID())
	}
	return false

}

//Connect begin the connect thread to given adderss
func (n *NetServer) Connect(addr string) {

}

//Halt stop all net layer logic
func (n *NetServer) Halt() {

}
