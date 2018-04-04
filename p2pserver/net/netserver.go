package netserver

import (
	"github.com/Ontology/common/config"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
	"net"
	"strconv"
	"strings"
)

type NetServer struct {
	Self        *peer.Peer
	ReceiveChan chan types.MsgPayload
}

func (n *NetServer) InitNonTLSListen() (net.Listener, net.Listener, error) {
	return nil, nil, nil
}
func (n *NetServer) InitTLSListen() (net.Listener, error) {
	return nil, nil
}
func (n *NetServer) NonTLSDial(addr string) (net.Conn, error) {
	return nil, nil
}
func (n *NetServer) TLSDial(addr string) (net.Conn, error) {
	return nil, nil
}
func (n *NetServer) Connectseeds() {
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
				n.ReqNeighborList(p)
			}
		} else { //not found
			go n.Connect(nodeAddr)
		}
	}
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
func (n *NetServer) GetId() uint64 {
	return n.Self.GetID()
}
func (n *NetServer) GetTime() int64 {
	return n.Self.GetTime()
}
func (n *NetServer) GetState() uint32 {
	return n.Self.GetState()
}
func (n *NetServer) GetServices() uint64 {
	return n.Self.GetServices()
}
func (n *NetServer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return n.Self.Np.GetNeighborAddrs()
}
func (n *NetServer) GetConnectionCnt() uint32 {
	return n.Self.Np.GetNbrNodeCnt()
}
func (n *NetServer) GetMsgCh() chan types.MsgPayload {
	return n.ReceiveChan
}
func (n *NetServer) Xmit(msg interface{}) error {
	return nil
}
func (n *NetServer) Tx(id uint64, data []byte) {
	node, ok := n.Self.Np.List[id]
	if ok == true {
		node.Send(data)
	}
}
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
func DisconnectNotify(v interface{}) {
	if p, ok := v.(*peer.Peer); ok {
		p.Close()
	}
}
func (n *NetServer) IsPeerEstablished(id uint64) bool {
	return n.Self.Np.NodeEstablished(id)
}
func (n *NetServer) ReqNeighborList(*peer.Peer) {

}
func (n *NetServer) Connect(addr string) {

}
func (n *NetServer) Halt() {

}
