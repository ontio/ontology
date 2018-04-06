package p2pserver

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actor "github.com/Ontology/p2pserver/actor/req"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

type P2PServer struct {
	Self          *peer.Peer
	network       P2P
	msgRouter     *MessageRouter
	quitOnline    chan bool
	quitHeartBeat chan bool
	quitSyncBlk   chan bool
}

func NewServer(acc *account.Account) (*P2PServer, error) {
	self, err := peer.NewPeer(acc.PubKey())
	if err != nil {
		return nil, err
	}
	n := NewNetServer(self)

	p := &P2PServer{
		Self:    self,
		network: n,
	}

	p.msgRouter = NewMsgRouter(p)

	// Fixme: implement the message handler for each msg type
	p.msgRouter.RegisterMsgHandler(types.VERSION_TYPE, VersionHandle)
	p.quitOnline = make(chan bool)
	p.quitHeartBeat = make(chan bool)
	p.quitSyncBlk = make(chan bool)
	return p, nil
}

func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}
func (this *P2PServer) Start(isSync bool) error {
	if this != nil {
		this.network.Start()
	}
	go this.keepOnline()
	go this.heartBeat()
	go this.syncBlock()
	return nil
}
func (this *P2PServer) Stop() error {
	this.network.Halt()
	return nil
}
func (this *P2PServer) IsSyncing() bool {
	return false
}
func (this *P2PServer) GetPort() (uint16, uint16) {
	return this.network.GetPort(), this.network.GetConsensusPort()
}
func (this *P2PServer) GetVersion() uint32 {
	return this.network.GetVersion()
}
func (this *P2PServer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return this.network.GetNeighborAddrs()
}
func (this *P2PServer) Xmit(msg interface{}) error {
	return nil
}
func (this *P2PServer) Send(id uint64, buf []byte, isConsensus bool) {
	if this.network.IsPeerEstablished(id) {
		this.network.Send(id, buf, isConsensus)
	}
	log.Errorf("P2PServer send error: peer %x is not established.", id)
}
func (this *P2PServer) GetId() uint64 {
	return this.network.GetId()
}
func (this *P2PServer) GetConnectionState() uint32 {
	return this.network.GetState()
}
func (this *P2PServer) GetTime() int64 {
	return this.network.GetTime()
}
func (this *P2PServer) blockSyncFinished() bool {
	peers := this.Self.Np.GetNeighbors()
	if len(peers) == 0 {
		return true
	}

	blockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("P2PServer GetCurrentBlockHeight error:%s", err)
		return false
	}

	for _, v := range peers {
		if blockHeight < uint32(v.GetHeight()) {
			return false
		}
	}
	return true
}

func (this *P2PServer) WaitForSyncBlkFinish() {
	for {
		headerHeight, _ := actor.GetCurrentHeaderHeight()
		currentBlkHeight, _ := actor.GetCurrentBlockHeight()
		log.Info("WaitForSyncBlkFinish... current block height is ", currentBlkHeight, " ,current header height is ", headerHeight)

		if this.blockSyncFinished() {
			break
		}
		<-time.After(types.PERIOD_UPDATE_TIME * time.Second)
	}
}

func (this *P2PServer) WaitForPeersStart() {
	for {
		log.Info("Wait for default connection...")
		if this.reachMinConnection() {
			break
		}
		<-time.After(types.PERIOD_UPDATE_TIME * time.Second)
	}
}

//connectSeeds connect the seeds in seedlist and call for nbr list
func (this *P2PServer) connectSeeds() {
	if this.reachMinConnection() {
		return
	}
	seedNodes := config.Parameters.SeedList
	for _, nodeAddr := range seedNodes {
		found := false
		var p *peer.Peer
		var ip net.IP
		this.Self.Np.Lock()
		for _, tn := range this.Self.Np.List {
			ipAddr, _ := tn.GetAddr16()
			ip = ipAddr[:]
			addrstring := ip.To16().String() + ":" + strconv.Itoa(int(tn.GetPort()))
			if nodeAddr == addrstring {
				p = tn
				found = true
				break
			}
		}
		this.Self.Np.Unlock()
		if found {
			if p.GetState() == types.ESTABLISH {
				this.reqNbrList(p)
			}
		} else { //not found
			go this.network.Connect(nodeAddr)
		}
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
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
	return int(this.GetConnectionCnt())+1 >= minCount
}

//retryInactivePeer try to connect peer in INACTIVITY state
func (this *P2PServer) retryInactivePeer() {

}

//keepOnline make sure seed peer be connected and try connect lost peer
func (this *P2PServer) keepOnline() {
	t := time.NewTimer(time.Second * types.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.connectSeeds()
			this.retryInactivePeer()
			t.Stop()
			t.Reset(time.Second * types.CONN_MONITOR)
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//reqNbrList ask the peer for its neighbor list
func (this *P2PServer) reqNbrList(*peer.Peer) {

}

//heartBeat send ping to nbr peers and check the timeout
func (this *P2PServer) heartBeat() {
	var periodTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodTime = config.Parameters.GenBlockTime / types.UPDATE_RATE_PER_BLOCK
	} else {
		periodTime = config.DEFAULTGENBLOCKTIME / types.UPDATE_RATE_PER_BLOCK
	}
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))

	for {
		select {
		case <-t.C:
			this.ping()
			this.timeout()
		case <-this.quitHeartBeat:
			t.Stop()
			break
		}
	}
}

//ping send pkg to get pong msg from others
func (this *P2PServer) ping() {

}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {

}

//syncBlock start sync up block
func (this *P2PServer) syncBlock() {

}
