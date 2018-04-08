package p2pserver

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/p2pserver/peer"
	actor "github.com/Ontology/p2pserver/actor/req"
	types "github.com/Ontology/p2pserver/common"
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
	//p.msgRouter.RegisterMsgHandler(types.VERSION_TYPE, VersionHandle)
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
	go this.keepOnlineService()
	go this.heartBeatService()
	go this.keepOnlineService()
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
func (this *P2PServer) Send(p *peer.Peer, buf []byte, isConsensus bool) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, buf, isConsensus)
	}
	log.Errorf("P2PServer send to a not ESTABLISH peer 0x%x", p.GetID())
	return errors.New("send to a not ESTABLISH peer")
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
		return false
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
			addrString := ip.To16().String() + ":" + strconv.Itoa(int(tn.GetSyncPort()))
			if nodeAddr == addrString {
				p = tn
				found = true
				break
			}
		}
		this.Self.Np.Unlock()
		if found {
			if p.GetSyncState() == types.ESTABLISH {
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
func (this *P2PServer) keepOnlineService() {
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
func (this *P2PServer) heartBeatService() {
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
	peers := this.Self.Np.GetNeighbors()
	for _, p := range peers {
		if p.GetSyncState() == types.ESTABLISH {
			height, err := actor.GetCurrentBlockHeight()
			if err != nil {
				log.Error("failed get current height! Ping faild!")
				return
			}
			buf, err := NewPingMsg(uint64(height))
			if err != nil {
				log.Error("failed build a new ping message")
			} else {
				go this.Send(p, buf, false)
			}
		}
	}
}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {

}

//syncBlock start sync up hdr & block
func (this *P2PServer) syncService() {
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
			this.syncBlockHdr()
			this.syncBlock()
		case <-this.quitHeartBeat:
			t.Stop()
			break
		}
	}
}

//syncBlockHdr send synchdr cmd to peers
func (this *P2PServer) syncBlockHdr() {
	if !this.reachMinConnection() {
		return
	}
	peers := this.Self.Np.GetNeighbors()
	if len(peers) == 0 {
		return
	}
	p := this.randPeer(peers)
	if p == nil {
		return
	}
	headerHash, _ := actor.GetCurrentHeaderHash()
	buf, err := NewHeadersReq(headerHash)
	if err != nil {
		log.Error("failed build a new headersReq")
	} else {
		go this.Send(p, buf, false)
	}
}

//syncBlock send reqblk cmd to peers
func (this *P2PServer) syncBlock() {
	headerHeight, _ := actor.GetCurrentHeaderHeight()
	currentBlkHeight, _ := actor.GetCurrentBlockHeight()
	if currentBlkHeight >= headerHeight {
		return
	}
	//TODO
	//var dValue int32
	//var reqCnt uint32
	//var i uint32
	//nodes := this.Self.Np.GetNeighbors()
	/*
		for _, n := range nodes {
			if uint32(n.GetHeight()) <= currentBlkHeight {
				continue
			}
			n.RemoveFlightHeightLessThan(currentBlkHeight)
			count := types.MAX_REQ_BLK_ONCE - uint32(n.GetFlightHeightCnt())
			dValue = int32(headerHeight - currentBlkHeight - reqCnt)
			flights := n.GetFlightHeights()
			if count == 0 {
				for _, f := range flights {
					hash, _ := actor.GetBlockHashByHeight(f)
					isContainBlock, _ := actor.IsContainBlock(hash)
					if isContainBlock == false {
						message.ReqBlkData(n, hash)
					}
				}

			}
			for i = 1; i <= count && dValue >= 0; i++ {
				hash, _ := actor.GetBlockHashByHeight(currentBlkHeight + reqCnt)
				isContainBlock, _ := actor.IsContainBlock(hash)
				if isContainBlock == false {
					message.ReqBlkData(n, hash)
					n.StoreFlightHeight(currentBlkHeight + reqCnt)
				}
				reqCnt++
				dValue--
			}
		}
	*/
}

//
func (this *P2PServer) randPeer(plist []*peer.Peer) *peer.Peer {
	selectList := []*peer.Peer{}
	for _, v := range plist {
		height, _ := actor.GetCurrentHeaderHeight()
		if uint64(height) < v.GetHeight() {
			selectList = append(selectList, v)
		}
	}
	nCount := len(selectList)
	if nCount == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(nCount)

	return selectList[index]
}
