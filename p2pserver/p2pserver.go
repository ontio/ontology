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
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgtypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols"
)

//P2PServer control all network activities
type P2PServer struct {
	network     p2pnet.P2P
	ledger      *ledger.Ledger
	recentPeers map[uint32][]string
	quit        chan bool
}

//NewServer return a new p2pserver according to the pubkey
func NewServer() (*P2PServer, error) {
	ld := ledger.DefLedger
	protocol := protocols.NewMsgHandler(ld)
	n, err := netserver.NewNetServer(protocol, config.DefConfig.P2PNode)
	if err != nil {
		return nil, err
	}

	p := &P2PServer{
		network: n,
		ledger:  ld,
	}

	p.loadRecentPeers()
	p.quit = make(chan bool)
	return p, nil
}

//Start create all services
func (this *P2PServer) Start() error {
	this.network.Start()
	this.tryRecentPeers()
	go this.connectSeedService()
	go this.syncUpRecentPeers()
	go this.heartBeatService()
	return nil
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() {
	this.network.Halt()
	this.quit <- true
}

// GetNetWork returns the low level netserver
func (this *P2PServer) GetNetWork() p2pnet.P2P {
	return this.network
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {
	log.Debug()
	var msg msgtypes.Message
	switch message.(type) {
	case *types.Transaction:
		log.Debug("[p2p]TX transaction message")
		txn := message.(*types.Transaction)
		msg = msgpack.NewTxn(txn)
	case *msgtypes.ConsensusPayload:
		log.Debug("[p2p]TX consensus message")
		consensusPayload := message.(*msgtypes.ConsensusPayload)
		msg = msgpack.NewConsensus(consensusPayload)
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
	this.network.Xmit(msg)
	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, msg msgtypes.Message) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, msg)
	}
	log.Warnf("[p2p]send to a not ESTABLISH peer %d",
		p.GetID())
	return errors.New("[p2p]send to a not ESTABLISH peer")
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

	connPeers := make(map[string]*peer.Peer)
	np := this.network.GetNp()
	np.Lock()
	for _, tn := range np.List {
		ipAddr, _ := tn.GetAddr16()
		ip := net.IP(ipAddr[:])
		addrString := ip.To16().String() + ":" + strconv.Itoa(int(tn.GetPort()))
		if tn.GetState() == common.ESTABLISH {
			connPeers[addrString] = tn
		}
	}
	np.Unlock()

	seedConnList := make([]*peer.Peer, 0)
	seedDisconn := make([]string, 0)
	isSeed := false
	for _, nodeAddr := range seedNodes {
		if p, ok := connPeers[nodeAddr]; ok {
			seedConnList = append(seedConnList, p)
		} else {
			seedDisconn = append(seedDisconn, nodeAddr)
		}

		if this.network.IsOwnAddress(nodeAddr) {
			isSeed = true
		}
	}

	if len(seedConnList) > 0 {
		rand.Seed(time.Now().UnixNano())
		// close NewAddrReq
		index := rand.Intn(len(seedConnList))
		this.reqNbrList(seedConnList[index])
		if isSeed && len(seedDisconn) > 0 {
			index := rand.Intn(len(seedDisconn))
			go this.network.Connect(seedDisconn[index])
		}
	} else { //not found
		for _, nodeAddr := range seedNodes {
			go this.network.Connect(nodeAddr)
		}
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) reachMinConnection() bool {
	if !config.DefConfig.Consensus.EnableConsensus {
		//just sync
		return true
	}
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
	return int(this.network.GetConnectionCnt())+1 >= minCount
}

//getNode returns the peer with the id
func (this *P2PServer) getNode(id uint64) *peer.Peer {
	return this.network.GetPeer(id)
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
		case <-this.quit:
			t.Stop()
			return
		}
	}
}

//reqNbrList ask the peer for its neighbor list
func (this *P2PServer) reqNbrList(p *peer.Peer) {
	id := p.GetKId()
	var msg msgtypes.Message
	if id.IsPseudoKadId() {
		msg = msgpack.NewAddrReq()
	} else {
		msg = msgpack.NewFindNodeReq(this.GetNetWork().GetKId())
	}

	go this.Send(p, msg)
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
		case <-this.quit:
			t.Stop()
			return
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
		if p.GetState() == common.ESTABLISH {
			height := this.ledger.GetCurrentBlockHeight()
			ping := msgpack.NewPingMsg(uint64(height))
			go this.Send(p, ping)
		}
	}
}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {
	peers := this.network.GetNeighbors()
	var periodTime uint
	periodTime = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for _, p := range peers {
		if p.GetState() == common.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second *
				time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
				log.Warnf("[p2p]keep alive timeout!!!lost remote peer %d - %s from %s", p.GetID(), p.Link.GetAddr(), t.String())
				p.Close()
			}
		}
	}
}

func (this *P2PServer) loadRecentPeers() {
	this.recentPeers = make(map[uint32][]string)
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
	}
}

//tryRecentPeers try connect recent contact peer when service start
func (this *P2PServer) tryRecentPeers() {
	netID := config.DefConfig.P2PNode.NetworkMagic
	if len(this.recentPeers[netID]) > 0 {
		log.Info("[p2p] try to connect recent peer")
	}
	for _, v := range this.recentPeers[netID] {
		go this.network.Connect(v)
	}
}

//syncUpRecentPeers sync up recent peers periodically
func (this *P2PServer) syncUpRecentPeers() {
	periodTime := common.RECENT_TIMEOUT
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))
	for {
		select {
		case <-t.C:
			this.persistRecentPeers()
		case <-this.quit:
			t.Stop()
			return
		}
	}
}

//persistRecentPeers compare snapshot of recent peer with current link,then persist the list
func (this *P2PServer) persistRecentPeers() {
	changed := false
	netID := config.DefConfig.P2PNode.NetworkMagic
	for i := 0; i < len(this.recentPeers[netID]); i++ {
		p := this.network.GetPeerFromAddr(this.recentPeers[netID][i])
		if p == nil || p.GetState() != common.ESTABLISH {
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
			nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(p.GetPort()))
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
