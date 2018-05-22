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

package netserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer(pubKey keypair.PublicKey) p2p.P2P {

	n := &NetServer{
		SyncChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
		ConsChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
	}

	n.PeerAddrMap.PeerSyncAddress = make(map[string]*peer.Peer)
	n.PeerAddrMap.PeerConsAddress = make(map[string]*peer.Peer)

	n.init(pubKey)
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base         peer.PeerCom
	synclistener net.Listener
	conslistener net.Listener
	SyncChan     chan *common.MsgPayload
	ConsChan     chan *common.MsgPayload
	ConnectingNodes
	PeerAddrMap
	Np *peer.NbrPeers
}

//ConnectingNodes include all addr in connecting state
type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerSyncAddress map[string]*peer.Peer
	PeerConsAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (this *NetServer) init(pubKey keypair.PublicKey) error {
	this.base.SetVersion(common.PROTOCOL_VERSION)

	if config.DefConfig.Consensus.EnableConsensus {
		this.base.SetServices(uint64(common.VERIFY_NODE))
	} else {
		this.base.SetServices(uint64(common.SERVICE_NODE))
	}

	if config.DefConfig.P2PNode.NodePort == 0 {
		log.Error("link port invalid")
		return errors.New("invalid link port")
	}

	this.base.SetSyncPort(uint16(config.DefConfig.P2PNode.NodePort))

	if config.DefConfig.P2PNode.DualPortSupport {
		if config.DefConfig.P2PNode.NodeConsensusPort == 0 {
			log.Error("consensus port invalid")
			return errors.New("invalid consensus port")
		}

		this.base.SetConsPort(uint16(config.DefConfig.P2PNode.NodeConsensusPort))
	} else {
		this.base.SetConsPort(0)
	}

	this.base.SetRelay(true)

	var id uint64
	key := keypair.SerializePublicKey(pubKey)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(id))
	if err != nil {
		log.Error(err)
		return err
	}
	this.base.SetID(id)

	log.Info(fmt.Sprintf("Init peer ID to 0x%x", this.base.GetID()))
	this.Np = &peer.NbrPeers{}
	this.Np.Init()

	this.base.SetPubKey(pubKey)
	return nil
}

//InitListen start listening on the config port
func (this *NetServer) Start() {
	this.startListening()
}

//GetVersion return self peer`s version
func (this *NetServer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetId return peer`s id
func (this *NetServer) GetID() uint64 {
	return this.base.GetID()
}

// SetHeight sets the local's height
func (this *NetServer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

// GetHeight return peer's heigh
func (this *NetServer) GetHeight() uint64 {
	return this.base.GetHeight()
}

//GetTime return the last contact time of self peer
func (this *NetServer) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

//GetServices return the service state of self peer
func (this *NetServer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetSyncPort return the sync port
func (this *NetServer) GetSyncPort() uint16 {
	return this.base.GetSyncPort()
}

//GetConsPort return the cons port
func (this *NetServer) GetConsPort() uint16 {
	return this.base.GetConsPort()
}

//GetHttpInfoPort return the port support info via http
func (this *NetServer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//GetRelay return whether net module can relay msg
func (this *NetServer) GetRelay() bool {
	return this.base.GetRelay()
}

//GetPubKey return the key config in net module
func (this *NetServer) GetPubKey() keypair.PublicKey {
	return this.base.GetPubKey()
}

// GetPeer returns a peer with the peer id
func (this *NetServer) GetPeer(id uint64) *peer.Peer {
	return this.Np.GetPeer(id)
}

//return nbr peers collection
func (this *NetServer) GetNp() *peer.NbrPeers {
	return this.Np
}

//GetNeighborAddrs return all the nbr peer`s addr
func (this *NetServer) GetNeighborAddrs() []common.PeerAddr {
	return this.Np.GetNeighborAddrs()
}

//GetConnectionCnt return the total number of valid connections
func (this *NetServer) GetConnectionCnt() uint32 {
	return this.Np.GetNbrNodeCnt()
}

//AddNbrNode add peer to nbr peer list
func (this *NetServer) AddNbrNode(remotePeer *peer.Peer) {
	this.Np.AddNbrNode(remotePeer)
}

//DelNbrNode delete nbr peer by id
func (this *NetServer) DelNbrNode(id uint64) (*peer.Peer, bool) {
	return this.Np.DelNbrNode(id)
}

//GetNeighbors return all nbr peer
func (this *NetServer) GetNeighbors() []*peer.Peer {
	return this.Np.GetNeighbors()
}

//NodeEstablished return whether a peer is establish with self according to id
func (this *NetServer) NodeEstablished(id uint64) bool {
	return this.Np.NodeEstablished(id)
}

//Xmit called by actor, broadcast msg
func (this *NetServer) Xmit(buf []byte, isCons bool) {
	this.Np.Broadcast(buf, isCons)
}

//GetMsgChan return sync or consensus channel when msgrouter need msg input
func (this *NetServer) GetMsgChan(isConsensus bool) chan *common.MsgPayload {
	if isConsensus {
		return this.ConsChan
	} else {
		return this.SyncChan
	}
}

//Tx send data buf to peer
func (this *NetServer) Send(p *peer.Peer, data []byte, isConsensus bool) error {
	if p != nil {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			return p.Send(data, false)
		}
		return p.Send(data, isConsensus)
	}
	log.Error("send to a invalid peer")
	return errors.New("send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (this *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return this.Np.NodeEstablished(p.GetID())
	}
	return false

}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) Connect(addr string, isConsensus bool) error {
	if this.IsNbrPeerAddr(addr, isConsensus) {
		return nil
	}
	if added := this.AddInConnectingList(addr); added == false {
		log.Info("node exist in connecting list", addr)
		return errors.New("node exist in connecting list")
	}

	isTls := config.DefConfig.P2PNode.IsTLS
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Error("connect failed: ", err)
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Error("connect failed: ", err)
			return err
		}
	}

	addr = conn.RemoteAddr().String()
	log.Info(fmt.Sprintf("peer %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))

	if !isConsensus {
		remotePeer = peer.NewPeer()
		this.AddPeerSyncAddress(addr, remotePeer)
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(this.SyncChan)
		go remotePeer.SyncLink.Rx()
		remotePeer.SetSyncState(common.HAND)
		vpl := msgpack.NewVersionPayload(this, false, ledger.DefLedger.GetCurrentBlockHeight())
		buf, _ := msgpack.NewVersion(vpl, this.GetPubKey())
		remotePeer.SyncLink.Tx(buf)
	} else {
		remotePeer = peer.NewPeer() //would merge with a exist peer in versionhandle
		this.AddPeerConsAddress(addr, remotePeer)
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetConsState(common.HAND)
		vpl := msgpack.NewVersionPayload(this, true, ledger.DefLedger.GetCurrentBlockHeight())
		buf, _ := msgpack.NewVersion(vpl, this.GetPubKey())
		remotePeer.ConsLink.Tx(buf)
	}

	return nil
}

//Halt stop all net layer logic
func (this *NetServer) Halt() {
	peers := this.Np.GetNeighbors()
	for _, p := range peers {
		p.CloseSync()
		p.CloseCons()
	}
	if this.synclistener != nil {
		this.synclistener.Close()
	}
	if this.conslistener != nil {
		this.conslistener.Close()
	}

}

//establishing the connection to remote peers and listening for inbound peers
func (this *NetServer) startListening() error {
	var err error

	syncPort := this.base.GetSyncPort()
	consPort := this.base.GetConsPort()

	if syncPort == 0 {
		log.Error("Sync Port invalid")
		return errors.New("Sync Port invalid")
	}

	err = this.startSyncListening(syncPort)
	if err != nil {
		return err
	}

	//consensus
	if config.DefConfig.P2PNode.DualPortSupport == false {
		log.Info("Dual port mode not supported,keep single link")
		return nil
	}
	if consPort == 0 || consPort == syncPort {
		//still work
		log.Error("Consensus Port invalid,keep single link")
	} else {
		err = this.startConsListening(consPort)
		if err != nil {
			return err
		}
	}
	return nil
}

// startSyncListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startSyncListening(port uint16) error {
	var err error
	this.synclistener, err = createListener(port)
	if err != nil {
		log.Error("failed to create sync listener")
		return errors.New("failed to create sync listener")
	}

	go this.startSyncAccept(this.synclistener)
	log.Infof("Start listen on sync port %d", port)
	return nil
}

// startConsListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startConsListening(port uint16) error {
	var err error
	this.conslistener, err = createListener(port)
	if err != nil {
		log.Error("failed to create cons listener")
		return errors.New("failed to create cons listener")
	}

	go this.startConsAccept(this.conslistener)
	log.Infof("Start listen on consensus port %d", port)
	return nil
}

//startSyncAccept accepts the sync connection from the inbound peer
func (this *NetServer) startSyncAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote sync node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddPeerSyncAddress(addr, remotePeer)
		if err != nil {
			log.Errorf("Error parse remote ip:%s", addr)
			return
		}
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(this.SyncChan)
		go remotePeer.SyncLink.Rx()
	}
}

//startConsAccept accepts the consensus connnection from the inbound peer
func (this *NetServer) startConsAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote cons node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddPeerConsAddress(addr, remotePeer)
		if err != nil {
			log.Errorf("Error parse remote ip:%s", addr)
			return
		}
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) AddInConnectingList(addr string) (added bool) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	for _, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	this.ConnectingAddrs = append(this.ConnectingAddrs, addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (this *NetServer) RemoveFromConnectingList(addr string) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	addrs := []string{}
	for i, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(this.ConnectingAddrs[:i], this.ConnectingAddrs[i+1:]...)
		}
	}
	this.ConnectingAddrs = addrs
}

//find exist peer from addr map
func (this *NetServer) GetPeerFromAddr(addr string) *peer.Peer {
	var p *peer.Peer
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()

	p, ok := this.PeerSyncAddress[addr]
	if ok {
		return p
	}
	p, ok = this.PeerConsAddress[addr]
	if ok {
		return p
	}
	return nil
}

//IsNbrPeerAddr return result whether the address is under connecting
func (this *NetServer) IsNbrPeerAddr(addr string, isConsensus bool) bool {
	var addrNew string
	this.Np.RLock()
	defer this.Np.RUnlock()
	for _, p := range this.Np.List {
		if p.GetSyncState() == common.HAND || p.GetSyncState() == common.HAND_SHAKE ||
			p.GetSyncState() == common.ESTABLISH {
			if isConsensus {
				addrNew = p.ConsLink.GetAddr()
			} else {
				addrNew = p.SyncLink.GetAddr()
			}
			if strings.Compare(addrNew, addr) == 0 {
				return true
			}
		}
	}
	return false
}

//AddPeerSyncAddress add sync addr to peer-addr map
func (this *NetServer) AddPeerSyncAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	this.PeerSyncAddress[addr] = p
}

//AddPeerConsAddress add cons addr to peer-addr map
func (this *NetServer) AddPeerConsAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	this.PeerConsAddress[addr] = p
}

//RemovePeerSyncAddress remove sync addr from peer-addr map
func (this *NetServer) RemovePeerSyncAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerSyncAddress[addr]; ok {
		delete(this.PeerSyncAddress, addr)
	}
}

//RemovePeerConsAddress remove cons addr from peer-addr map
func (this *NetServer) RemovePeerConsAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerConsAddress[addr]; ok {
		delete(this.PeerConsAddress, addr)
	}
}
