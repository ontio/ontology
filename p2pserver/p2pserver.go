/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package p2pserver

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actor "github.com/Ontology/p2pserver/actor/req"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

type P2PServer struct {
	Self      *peer.Peer
	network   P2P
	msgRouter *MessageRouter
	ReconnectAddrs
	flightHeights map[uint64][]uint32
	quitOnline    chan bool
	quitHeartBeat chan bool
	quitSyncBlk   chan bool
}

//reconnectAddrs contain addr need to reconnect
type ReconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//NewServer return a new p2pserver according to the pubkey
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

	// Register message handler
	p.msgRouter.RegisterMsgHandler(types.VERSION_TYPE, VersionHandle)
	p.msgRouter.RegisterMsgHandler(types.VERACK_TYPE, VerAckHandle)
	p.msgRouter.RegisterMsgHandler(types.GetADDR_TYPE, AddrReqHandle)
	p.msgRouter.RegisterMsgHandler(types.ADDR_TYPE, AddrHandle)
	p.msgRouter.RegisterMsgHandler(types.PING_TYPE, PingHandle)
	p.msgRouter.RegisterMsgHandler(types.PONG_TYPE, PongHandle)
	p.msgRouter.RegisterMsgHandler(types.GET_HEADERS_TYPE, HeadersReqHandle)
	p.msgRouter.RegisterMsgHandler(types.HEADERS_TYPE, BlkHeaderHandle)
	p.msgRouter.RegisterMsgHandler(types.GET_BLOCKS_TYPE, BlocksReqHandle)
	p.msgRouter.RegisterMsgHandler(types.INV_TYPE, InvHandle)
	p.msgRouter.RegisterMsgHandler(types.GET_DATA_TYPE, DataReqHandle)
	p.msgRouter.RegisterMsgHandler(types.BLOCK_TYPE, BlockHandle)
	p.msgRouter.RegisterMsgHandler(types.CONSENSUS_TYPE, ConsensusHandle)
	p.msgRouter.RegisterMsgHandler(types.NOT_FOUND_TYPE, NotFoundHandle)
	p.msgRouter.RegisterMsgHandler(types.TX_TYPE, TransactionHandle)

	//p.msgRouter.RegisterMsgHandler(types.VERSION_TYPE, VersionHandle)
	p.flightHeights = make(map[uint64][]uint32)
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
	this.quitOnline <- true
	this.quitHeartBeat <- true
	this.quitSyncBlk <- true

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
	this.Self.Np.Lock()
	var ip net.IP
	neighborPeers := make(map[uint64]*peer.Peer)
	for _, p := range this.Self.Np.List {
		addr, _ := p.GetAddr16()
		ip = addr[:]
		nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(p.GetSyncPort()))
		if p.GetSyncState() == types.INACTIVITY {
			//add addr to retry list
			this.addToRetryList(nodeAddr)
			//close legacy node
			p.CloseSync()
			p.CloseCons()
		} else {
			//add others to tmp node map
			this.removeFromRetryList(nodeAddr)
			neighborPeers[p.GetID()] = p
		}
	}

	this.Self.Np.List = neighborPeers
	this.Self.Np.Unlock()
	//try connect
	if len(this.RetryAddrs) > 0 {
		this.ReconnectAddrs.Lock()

		list := make(map[string]int)
		for addr := range this.RetryAddrs {
			this.RetryAddrs[addr] = this.RetryAddrs[addr] + 1
			rand.Seed(time.Now().UnixNano())
			log.Trace("Try to reconnect peer, peer addr is ", addr)
			<-time.After(time.Duration(rand.Intn(types.CONN_MAX_BACK)) * time.Millisecond)
			log.Trace("Back off time`s up, start connect node")
			this.network.Connect(addr)
			if this.RetryAddrs[addr] < types.MAX_RETRY_COUNT {
				list[addr] = this.RetryAddrs[addr]
			}
		}
		this.RetryAddrs = list
		this.ReconnectAddrs.Unlock()
	}
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
func (this *P2PServer) reqNbrList(p *peer.Peer) {
	buf, _ := NewAddrReq()
	go this.Send(p, buf, false)
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
	peers := this.Self.Np.GetNeighbors()
	var periodTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodTime = config.Parameters.GenBlockTime / types.UPDATE_RATE_PER_BLOCK
	} else {
		periodTime = config.DEFAULTGENBLOCKTIME / types.UPDATE_RATE_PER_BLOCK
	}
	for _, p := range peers {
		if p.GetSyncState() == types.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second * time.Duration(periodTime) * types.KEEPALIVE_TIMEOUT)) {
				log.Warn("Keep alive timeout!!!")
				p.SetSyncState(types.INACTIVITY)
				p.SetConsState(types.INACTIVITY)
				p.CloseSync()
				p.CloseCons()
			}
		}
	}
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
	p := randPeer(peers)
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
	var dValue int32
	var reqCnt uint32

	currentHdrHeight, _ := actor.GetCurrentHeaderHeight()
	currentBlkHeight, _ := actor.GetCurrentBlockHeight()
	if currentBlkHeight >= currentHdrHeight {
		return
	}

	peers := this.Self.Np.GetNeighbors()

	for _, p := range peers {
		if uint32(p.GetHeight()) <= currentBlkHeight {
			continue
		}
		this.removeFlightHeightLessThan(p, currentBlkHeight)
		count := types.MAX_REQ_BLK_ONCE - uint32(len(this.flightHeights[p.GetID()]))
		dValue = int32(currentHdrHeight - currentBlkHeight - reqCnt)
		flights := this.flightHeights[p.GetID()]
		if count == 0 {
			for _, f := range flights {
				hash, _ := actor.GetBlockHashByHeight(f)
				isContainBlock, _ := actor.IsContainBlock(hash)
				if isContainBlock == false {
					reqBuf, err := NewBlkDataReq(hash)
					if err != nil {
						log.Error("syncBlock error:", err)
						break
					}
					err = this.Send(p, reqBuf, false)
					if err != nil {
						log.Error("Send error:", err)
						break
					}
				}
			}

		} else {
			for i := uint32(1); i <= count && dValue >= 0; i++ {
				hash, _ := actor.GetBlockHashByHeight(currentBlkHeight + reqCnt)
				isContainBlock, _ := actor.IsContainBlock(hash)
				if isContainBlock == false {
					reqBuf, err := NewBlkDataReq(hash)
					if err != nil {
						log.Error("syncBlock error:", err)
					}
					err = this.Send(p, reqBuf, false)
					if err != nil {
						log.Error("Send error:", err)
						break
					}
					//store the flghtheight
					flights = append(flights, currentBlkHeight+reqCnt)
				}
				reqCnt++
				dValue--
			}
		}
		this.flightHeights[p.GetID()] = flights
	}

}

//removeFlightHeightLessThan remove peer`s flightheight less than given height
func (this *P2PServer) removeFlightHeightLessThan(p *peer.Peer, h uint32) {
	heights := this.flightHeights[p.GetID()]
	nCnt := len(heights)
	i := 0

	for i < nCnt {
		if heights[i] < h {
			nCnt--
			heights[nCnt], heights[i] = heights[i], heights[nCnt]
		} else {
			i++
		}
	}
	this.flightHeights[p.GetID()] = heights[:nCnt]
}

//RemoveFlightHeight remove given height in flights
func (this *P2PServer) RemoveFlightHeight(p *peer.Peer, height uint32) []uint32 {
	heights := this.flightHeights[p.GetID()]
	for i, v := range heights {
		if v == height {
			return append(heights[:i], heights[i+1:]...)
		}
	}
	return heights
}

//randPeer choose a random peer from given peers
func randPeer(plist []*peer.Peer) *peer.Peer {
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

//addToRetryList add retry address to ReconnectAddrs
func (this *P2PServer) addToRetryList(addr string) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if this.RetryAddrs == nil {
		this.RetryAddrs = make(map[string]int)
	}
	if _, ok := this.RetryAddrs[addr]; ok {
		delete(this.RetryAddrs, addr)
	}
	//alway set retry to 0
	this.RetryAddrs[addr] = 0
}

//removeFromRetryList remove connected address from ReconnectAddrs
func (this *P2PServer) removeFromRetryList(addr string) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if len(this.RetryAddrs) > 0 {
		if _, ok := this.RetryAddrs[addr]; ok {
			delete(this.RetryAddrs, addr)
		}
	}
}
