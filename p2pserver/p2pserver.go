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
	"bytes"
	"errors"
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
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
	ReconnectAddrs
	quitOnline    chan bool
	quitHeartBeat chan bool
}

//ReconnectAddrs contain addr need to reconnect
type ReconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]int
}

//NewServer return a new p2pserver according to the pubkey
func NewServer(acc *account.Account) (*P2PServer, error) {
	n := netserver.NewNetServer(acc.PubKey())

	p := &P2PServer{
		network: n,
	}

	p.msgRouter = utils.NewMsgRouter(p.network)
	p.blockSync = NewBlockSyncMgr(p)
	p.quitOnline = make(chan bool)
	p.quitHeartBeat = make(chan bool)
	return p, nil
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}

//Start create all services
func (this *P2PServer) Start() error {
	if this.network != nil {
		this.network.Start()
	}
	if this.msgRouter != nil {
		this.msgRouter.Start()
	}
	go this.keepOnlineService()
	go this.heartBeatService()
	go this.blockSync.Start()

	return nil
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() error {
	this.network.Halt()
	this.quitOnline <- true
	this.quitHeartBeat <- true
	this.msgRouter.Stop()
	this.blockSync.Close()
	return nil
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
func (this *P2PServer) GetNeighborAddrs() ([]common.PeerAddr, uint64) {
	return this.network.GetNeighborAddrs()
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {
	log.Debug()
	var buffer []byte
	var err error
	isConsensus := false
	switch message.(type) {
	case *types.Transaction:
		log.Debug("TX transaction message")
		txn := message.(*types.Transaction)
		buffer, err = msgpack.NewTxn(txn)
		if err != nil {
			log.Error("Error New Tx message: ", err)
			return err
		}

	case *types.Block:
		log.Debug("TX block message")
		block := message.(*types.Block)
		buffer, err = msgpack.NewBlock(block)
		if err != nil {
			log.Error("Error New Block message: ", err)
			return err
		}
	case *msgtypes.ConsensusPayload:
		log.Debug("TX consensus message")
		consensusPayload := message.(*msgtypes.ConsensusPayload)
		buffer, err = msgpack.NewConsensus(consensusPayload)
		if err != nil {
			log.Error("Error New consensus message: ", err)
			return err
		}
		isConsensus = true
	case comm.Uint256:
		log.Debug("TX block hash message")
		hash := message.(comm.Uint256)
		buf := bytes.NewBuffer([]byte{})
		hash.Serialize(buf)
		// construct inv message
		invPayload := msgpack.NewInvPayload(comm.BLOCK, 1, buf.Bytes())
		buffer, err = msgpack.NewInv(invPayload)
		if err != nil {
			log.Error("Error New inv message")
			return err
		}
	default:
		log.Warnf("Unknown Xmit message %v , type %v", message,
			reflect.TypeOf(message))
		return errors.New("Unknown Xmit message type")
	}
	this.network.Xmit(buffer, isConsensus)
	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, buf []byte,
	isConsensus bool) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, buf, isConsensus)
	}
	log.Errorf("P2PServer send to a not ESTABLISH peer 0x%x",
		p.GetID())
	return errors.New("send to a not ESTABLISH peer")
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
func (this *P2PServer) OnHeaderReceive(headers []*types.Header) {
	this.blockSync.OnHeaderReceive(headers)
}

// OnBlockReceive adds the block from network
func (this *P2PServer) OnBlockReceive(block *types.Block) {
	this.blockSync.OnBlockReceive(block)
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

//WaitForSyncBlkFinish compare the height of self and remote peer in loop
func (this *P2PServer) WaitForSyncBlkFinish() {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "solo" {
		return
	}

	for {
		headerHeight, _ := actor.GetCurrentHeaderHeight()
		currentBlkHeight, _ := actor.GetCurrentBlockHeight()
		log.Info("WaitForSyncBlkFinish... current block height is ",
			currentBlkHeight, " ,current header height is ", headerHeight)

		if this.blockSyncFinished() {
			break
		}

		<-time.After(time.Second * (time.Duration(common.SYNC_BLK_WAIT)))
	}
}

//WaitForPeersStart check whether enough peer linked in loop
func (this *P2PServer) WaitForPeersStart() {
	var periodTime uint
	for {
		log.Info("Wait for minimum connection...")
		if this.reachMinConnection() {
			break
		}
		if config.Parameters.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
			periodTime = config.Parameters.GenBlockTime / common.UPDATE_RATE_PER_BLOCK
		} else {
			periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
		}
		<-time.After(time.Second * (time.Duration(periodTime)))
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
		np := this.network.GetNp()
		np.Lock()
		for _, tn := range np.List {
			ipAddr, _ := tn.GetAddr16()
			ip = ipAddr[:]
			addrString := ip.To16().String() + ":" +
				strconv.Itoa(int(tn.GetSyncPort()))
			if nodeAddr == addrString {
				p = tn
				found = true
				break
			}
		}
		np.Unlock()
		if found {
			if p.GetSyncState() == common.ESTABLISH {
				this.reqNbrList(p)
			}
		} else { //not found
			go this.network.Connect(nodeAddr, false)
		}
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
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
			log.Infof(" try reconnect %s", nodeAddr)
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
		}
		this.RetryAddrs = list
		this.ReconnectAddrs.Unlock()
		for _, addr := range addrs {
			rand.Seed(time.Now().UnixNano())
			log.Info("Try to reconnect peer, peer addr is ", addr)
			<-time.After(time.Duration(rand.Intn(common.CONN_MAX_BACK)) * time.Millisecond)
			log.Info("Back off time`s up, start connect node")
			this.network.Connect(addr, false)
		}

	}
}

//keepOnline make sure seed peer be connected and try connect lost peer
func (this *P2PServer) keepOnlineService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.connectSeeds()
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
	buf, _ := msgpack.NewAddrReq()
	go this.Send(p, buf, false)
}

//heartBeat send ping to nbr peers and check the timeout
func (this *P2PServer) heartBeatService() {
	var periodTime uint
	if config.Parameters.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
		periodTime = config.Parameters.GenBlockTime / common.UPDATE_RATE_PER_BLOCK
	} else {
		periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
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
	peers := this.network.GetNeighbors()
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			height, err := actor.GetCurrentBlockHeight()
			if err != nil {
				log.Error("failed get current height! Ping faild!")
				return
			}
			buf, err := msgpack.NewPingMsg(uint64(height))
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
	peers := this.network.GetNeighbors()
	var periodTime uint
	if config.Parameters.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
		periodTime = config.Parameters.GenBlockTime / common.UPDATE_RATE_PER_BLOCK
	} else {
		periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	}
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second *
				time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
				log.Warn("Keep alive timeout!!!")
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
