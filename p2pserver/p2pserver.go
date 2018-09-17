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
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgtypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/message/utils"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//P2PServer control all network activities
type P2PServer struct {
	network   p2pnet.P2P
	msgRouter *utils.MessageRouter
	pid       *evtActor.PID
	blockSync *BlockSyncMgr
	ledger    *ledger.Ledger
	ReconnectAddrs
	recentPeers    map[uint32][]string
	quitSyncRecent chan bool
	quitOnline     chan bool
	quitHeartBeat  chan bool
}

//ReconnectAddrs contain addr need to reconnect
type ReconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//NewServer return a new p2pserver according to the pubkey
func NewServer() *P2PServer {
	n := netserver.NewNetServer()

	p := &P2PServer{
		network: n,
		ledger:  ledger.DefLedger,
	}

	p.msgRouter = utils.NewMsgRouter(p.network)
	p.blockSync = NewBlockSyncMgr(p)
	p.recentPeers = make(map[uint32][]string)
	p.quitSyncRecent = make(chan bool)
	p.quitOnline = make(chan bool)
	p.quitHeartBeat = make(chan bool)
	return p
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}

//Start create all services
func (this *P2PServer) Start() error {
	if this.network != nil {
		this.network.Start()
	} else {
		return errors.New("[p2p]network invalid")
	}
	if this.msgRouter != nil {
		this.msgRouter.Start()
	} else {
		return errors.New("[p2p]msg router invalid")
	}
	this.tryRecentPeers()
	go this.connectSeedService()
	go this.syncUpRecentPeers()
	go this.keepOnlineService()
	go this.heartBeatService()
	go this.blockSync.Start()
	return nil
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() {
	this.network.Halt()
	this.quitSyncRecent <- true
	this.quitOnline <- true
	this.quitHeartBeat <- true
	this.msgRouter.Stop()
	this.blockSync.Close()
}

// GetNetWork returns the low level netserver
func (this *P2PServer) GetNetWork() p2pnet.P2P {
	return this.network
}

//GetPort return two network port
func (this *P2PServer) GetPort() (uint16, uint16) {
	return this.network.GetSyncPort(), this.network.GetConsPort()
}

//GetVersion return self version
func (this *P2PServer) GetVersion() uint32 {
	return this.network.GetVersion()
}

//GetNeighborAddrs return all nbr`s address
func (this *P2PServer) GetNeighborAddrs() []common.PeerAddr {
	return this.network.GetNeighborAddrs()
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {
	log.Debug()
	var msg msgtypes.Message
	isConsensus := false
	switch message.(type) {
	case *types.Transaction:
		log.Debug("[p2p]TX transaction message")
		txn := message.(*types.Transaction)
		msg = msgpack.NewTxn(txn)
	case *types.Block:
		log.Debug("[p2p]TX block message")
		block := message.(*types.Block)
		msg = msgpack.NewBlock(block)
	case *msgtypes.ConsensusPayload:
		log.Debug("[p2p]TX consensus message")
		consensusPayload := message.(*msgtypes.ConsensusPayload)
		msg = msgpack.NewConsensus(consensusPayload)
		isConsensus = true
	case comm.Uint256:
		log.Debug("[p2p]TX block hash message")
		hash := message.(comm.Uint256)
		// construct inv message
		invPayload := msgpack.NewInvPayload(comm.BLOCK, []comm.Uint256{hash})
		msg = msgpack.NewInv(invPayload)
	default:
		log.Warnf("[p2p]Unknown Xmit message %v , type %v", message,
			reflect.TypeOf(message))
		return errors.New("[p2p]Unknown Xmit message type")
	}
	this.network.Xmit(msg, isConsensus)
	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, msg msgtypes.Message,
	isConsensus bool) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, msg, isConsensus)
	}
	log.Warnf("[p2p]send to a not ESTABLISH peer %d",
		p.GetID())
	return errors.New("[p2p]send to a not ESTABLISH peer")
}

// GetID returns local node id
func (this *P2PServer) GetID() uint64 {
	return this.network.GetID()
}

// OnAddNode adds the peer id to the block sync mgr
func (this *P2PServer) OnAddNode(id uint64) {
	this.blockSync.OnAddNode(id)
}

// OnDelNode removes the peer id from the block sync mgr
func (this *P2PServer) OnDelNode(id uint64) {
	this.blockSync.OnDelNode(id)
}

// OnHeaderReceive adds the header list from network
func (this *P2PServer) OnHeaderReceive(fromID uint64, headers []*types.Header) {
	this.blockSync.OnHeaderReceive(fromID, headers)
}

// OnBlockReceive adds the block from network
func (this *P2PServer) OnBlockReceive(fromID uint64, blockSize uint32, block *types.Block) {
	this.blockSync.OnBlockReceive(fromID, blockSize, block)
}

// Todo: remove it if no use
func (this *P2PServer) GetConnectionState() uint32 {
	return common.INIT
}

//GetTime return lastet contact time
func (this *P2PServer) GetTime() int64 {
	return this.network.GetTime()
}

// SetPID sets p2p actor
func (this *P2PServer) SetPID(pid *evtActor.PID) {
	this.pid = pid
	this.msgRouter.SetPID(pid)
}

// GetPID returns p2p actor
func (this *P2PServer) GetPID() *evtActor.PID {
	return this.pid
}

//blockSyncFinished compare all nbr peers and self height at beginning
func (this *P2PServer) blockSyncFinished() bool {
	peers := this.network.GetNeighbors()
	if len(peers) == 0 {
		return false
	}

	blockHeight := this.ledger.GetCurrentBlockHeight()

	for _, v := range peers {
		if blockHeight < uint32(v.GetHeight()) {
			return false
		}
	}
	return true
}

//WaitForSyncBlkFinish compare the height of self and remote peer in loop
func (this *P2PServer) WaitForSyncBlkFinish() {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "solo" {
		return
	}

	for {
		headerHeight := this.ledger.GetCurrentHeaderHeight()
		currentBlkHeight := this.ledger.GetCurrentBlockHeight()
		log.Info("[p2p]WaitForSyncBlkFinish... current block height is ",
			currentBlkHeight, " ,current header height is ", headerHeight)

		if this.blockSyncFinished() {
			break
		}

		<-time.After(time.Second * (time.Duration(common.SYNC_BLK_WAIT)))
	}
}

//WaitForPeersStart check whether enough peer linked in loop
func (this *P2PServer) WaitForPeersStart() {
	periodTime := config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for {
		log.Info("[p2p]Wait for minimum connection...")
		if this.reachMinConnection() {
			break
		}

		<-time.After(time.Second * (time.Duration(periodTime)))
	}
}

//connectSeeds connect the seeds in seedlist and call for nbr list
func (this *P2PServer) connectSeeds() {
	seedNodes := make([]string, 0)
	pList := make([]*peer.Peer, 0)
	for _, n := range config.DefConfig.Genesis.SeedList {
		ip, err := common.ParseIPAddr(n)
		if err != nil {
			log.Warnf("[p2p]seed peer %s address format is wrong", n)
			continue
		}
		ns, err := net.LookupHost(ip)
		if err != nil {
			log.Warnf("[p2p]resolve err: %s", err.Error())
			continue
		}
		port, err := common.ParseIPPort(n)
		if err != nil {
			log.Warnf("[p2p]seed peer %s address format is wrong", n)
			continue
		}
		seedNodes = append(seedNodes, ns[0]+port)
	}

	for _, nodeAddr := range seedNodes {
		var ip net.IP
		np := this.network.GetNp()
		np.Lock()
		for _, tn := range np.List {
			ipAddr, _ := tn.GetAddr16()
			ip = ipAddr[:]
			addrString := ip.To16().String() + ":" +
				strconv.Itoa(int(tn.GetSyncPort()))
			if nodeAddr == addrString && tn.GetSyncState() == common.ESTABLISH {
				pList = append(pList, tn)
			}
		}
		np.Unlock()
	}
	if len(pList) > 0 {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(pList))
		this.reqNbrList(pList[index])
	} else { //not found
		for _, nodeAddr := range seedNodes {
			go this.network.Connect(nodeAddr, false)
		}
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFT_MIN_NODE_NUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLO_MIN_NODE_NUM
	case "vbft":
		minCount = config.VBFT_MIN_NODE_NUM

	}
	return int(this.GetConnectionCnt())+1 >= minCount
}

//getNode returns the peer with the id
func (this *P2PServer) getNode(id uint64) *peer.Peer {
	return this.network.GetPeer(id)
}

//retryInactivePeer try to connect peer in INACTIVITY state
func (this *P2PServer) retryInactivePeer() {
	np := this.network.GetNp()
	np.Lock()
	var ip net.IP
	neighborPeers := make(map[uint64]*peer.Peer)
	for _, p := range np.List {
		addr, _ := p.GetAddr16()
		ip = addr[:]
		nodeAddr := ip.To16().String() + ":" +
			strconv.Itoa(int(p.GetSyncPort()))
		if p.GetSyncState() == common.INACTIVITY {
			log.Debugf("[p2p] try reconnect %s", nodeAddr)
			//add addr to retry list
			this.addToRetryList(nodeAddr)
			p.CloseSync()
			p.CloseCons()
		} else {
			//add others to tmp node map
			this.removeFromRetryList(nodeAddr)
			neighborPeers[p.GetID()] = p
		}
	}

	np.List = neighborPeers
	np.Unlock()

	connCount := uint(this.network.GetOutConnRecordLen())
	if connCount >= config.DefConfig.P2PNode.MaxConnOutBound {
		log.Warnf("[p2p]Connect: out connections(%d) reach the max limit(%d)", connCount,
			config.DefConfig.P2PNode.MaxConnOutBound)
		return
	}

	//try connect
	if len(this.RetryAddrs) > 0 {
		this.ReconnectAddrs.Lock()

		list := make(map[string]int)
		addrs := make([]string, 0, len(this.RetryAddrs))
		for addr, v := range this.RetryAddrs {
			v += 1
			addrs = append(addrs, addr)
			if v < common.MAX_RETRY_COUNT {
				list[addr] = v
			}
			if v >= common.MAX_RETRY_COUNT {
				this.network.RemoveFromConnectingList(addr)
				remotePeer := this.network.GetPeerFromAddr(addr)
				if remotePeer != nil {
					if remotePeer.SyncLink.GetAddr() == addr {
						this.network.RemovePeerSyncAddress(addr)
						this.network.RemovePeerConsAddress(addr)
					}
					if remotePeer.ConsLink.GetAddr() == addr {
						this.network.RemovePeerConsAddress(addr)
					}
					this.network.DelNbrNode(remotePeer.GetID())
				}
			}
		}

		this.RetryAddrs = list
		this.ReconnectAddrs.Unlock()
		for _, addr := range addrs {
			rand.Seed(time.Now().UnixNano())
			log.Debug("[p2p]Try to reconnect peer, peer addr is ", addr)
			<-time.After(time.Duration(rand.Intn(common.CONN_MAX_BACK)) * time.Millisecond)
			log.Debug("[p2p]Back off time`s up, start connect node")
			this.network.Connect(addr, false)
		}

	}
}

//connectSeedService make sure seed peer be connected
func (this *P2PServer) connectSeedService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.connectSeeds()
			t.Stop()
			if this.reachMinConnection() {
				t.Reset(time.Second * time.Duration(10*common.CONN_MONITOR))
			} else {
				t.Reset(time.Second * common.CONN_MONITOR)
			}
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//keepOnline try connect lost peer
func (this *P2PServer) keepOnlineService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.retryInactivePeer()
			t.Stop()
			t.Reset(time.Second * common.CONN_MONITOR)
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//reqNbrList ask the peer for its neighbor list
func (this *P2PServer) reqNbrList(p *peer.Peer) {
	msg := msgpack.NewAddrReq()
	go this.Send(p, msg, false)
}

//heartBeat send ping to nbr peers and check the timeout
func (this *P2PServer) heartBeatService() {
	var periodTime uint
	periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
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
	peers := this.network.GetNeighbors()
	this.pingTo(peers)
}

//pings send pkgs to get pong msg from others
func (this *P2PServer) pingTo(peers []*peer.Peer) {
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			height := this.ledger.GetCurrentBlockHeight()
			ping := msgpack.NewPingMsg(uint64(height))
			go this.Send(p, ping, false)
		}
	}
}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {
	peers := this.network.GetNeighbors()
	var periodTime uint
	periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second *
				time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
				log.Warnf("[p2p]keep alive timeout!!!lost remote peer %d - %s from %s", p.GetID(), p.SyncLink.GetAddr(), t.String())
				p.CloseSync()
				p.CloseCons()
			}
		}
	}
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

//tryRecentPeers try connect recent contact peer when service start
func (this *P2PServer) tryRecentPeers() {
	netID := config.DefConfig.P2PNode.NetworkMagic
	if comm.FileExisted(common.RECENT_FILE_NAME) {
		buf, err := ioutil.ReadFile(common.RECENT_FILE_NAME)
		if err != nil {
			log.Warn("[p2p]read %s fail:%s, connect recent peers cancel", common.RECENT_FILE_NAME, err.Error())
			return
		}

		err = json.Unmarshal(buf, &this.recentPeers)
		if err != nil {
			log.Warn("[p2p]parse recent peer file fail: ", err)
			return
		}
		if len(this.recentPeers[netID]) > 0 {
			log.Info("[p2p]try to connect recent peer")
		}
		for _, v := range this.recentPeers[netID] {
			go this.network.Connect(v, false)
		}

	}
}

//syncUpRecentPeers sync up recent peers periodically
func (this *P2PServer) syncUpRecentPeers() {
	periodTime := common.RECENT_TIMEOUT
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))
	for {
		select {
		case <-t.C:
			this.syncPeerAddr()
		case <-this.quitSyncRecent:
			t.Stop()
			break
		}
	}

}

//syncPeerAddr compare snapshot of recent peer with current link,then persist the list
func (this *P2PServer) syncPeerAddr() {
	changed := false
	netID := config.DefConfig.P2PNode.NetworkMagic
	for i := 0; i < len(this.recentPeers[netID]); i++ {
		p := this.network.GetPeerFromAddr(this.recentPeers[netID][i])
		if p == nil || (p != nil && p.GetSyncState() != common.ESTABLISH) {
			this.recentPeers[netID] = append(this.recentPeers[netID][:i], this.recentPeers[netID][i+1:]...)
			changed = true
			i--
		}
	}
	left := common.RECENT_LIMIT - len(this.recentPeers[netID])
	if left > 0 {
		np := this.network.GetNp()
		np.Lock()
		var ip net.IP
		for _, p := range np.List {
			addr, _ := p.GetAddr16()
			ip = addr[:]
			nodeAddr := ip.To16().String() + ":" +
				strconv.Itoa(int(p.GetSyncPort()))
			found := false
			for i := 0; i < len(this.recentPeers[netID]); i++ {
				if nodeAddr == this.recentPeers[netID][i] {
					found = true
					break
				}
			}
			if !found {
				this.recentPeers[netID] = append(this.recentPeers[netID], nodeAddr)
				left--
				changed = true
				if left == 0 {
					break
				}
			}
		}
		np.Unlock()
	} else {
		if left < 0 {
			left = -left
			this.recentPeers[netID] = append(this.recentPeers[netID][:0], this.recentPeers[netID][0+left:]...)
			changed = true
		}

	}
	if changed {
		buf, err := json.Marshal(this.recentPeers)
		if err != nil {
			log.Warn("[p2p]package recent peer fail: ", err)
			return
		}
		err = ioutil.WriteFile(common.RECENT_FILE_NAME, buf, os.ModePerm)
		if err != nil {
			log.Warn("[p2p]write recent peer fail: ", err)
		}
	}
}
