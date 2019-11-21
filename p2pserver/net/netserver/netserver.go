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
	"errors"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/common/set"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer() p2p.P2P {
	n := &NetServer{
		NetChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
	}

	n.PeerAddrMap.PeerAddress = make(map[string]*peer.Peer)

	n.init()
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base     peer.PeerCom
	listener net.Listener
	NetChan  chan *types.MsgPayload
	connectingNodes
	PeerAddrMap
	Np            *peer.NbrPeers
	connectLock   sync.Mutex
	inConnRecord  InConnectionRecord
	outConnRecord OutConnectionRecord
	OwnAddress    string //network`s own address(ip : sync port),which get from version check
}

//InConnectionRecord include all addr connected
type InConnectionRecord struct {
	sync.RWMutex
	InConnectingAddrs set.StringSet
}

//OutConnectionRecord include all addr accepted
type OutConnectionRecord struct {
	sync.RWMutex
	OutConnectingAddrs set.StringSet
}

//connectingNodes include all addr in connecting state
type connectingNodes struct {
	sync.RWMutex
	ConnectingAddrs set.StringSet
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (this *NetServer) init() error {
	this.base.SetVersion(common.PROTOCOL_VERSION)

	if config.DefConfig.Consensus.EnableConsensus {
		this.base.SetServices(uint64(common.VERIFY_NODE))
	} else {
		this.base.SetServices(uint64(common.SERVICE_NODE))
	}

	if config.DefConfig.P2PNode.NodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}

	this.base.SetPort(uint16(config.DefConfig.P2PNode.NodePort))

	this.base.SetRelay(true)

	rand.Seed(time.Now().UnixNano())
	id := rand.Uint64()

	this.base.SetID(id)

	log.Infof("[p2p]init peer ID to %d", this.base.GetID())
	this.Np = &peer.NbrPeers{}
	this.Np.Init()

	this.connectingNodes.ConnectingAddrs = set.NewStringSet()
	this.inConnRecord.InConnectingAddrs = set.NewStringSet()
	this.outConnRecord.OutConnectingAddrs = set.NewStringSet()

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

//GetPort return the sync port
func (this *NetServer) GetPort() uint16 {
	return this.base.GetPort()
}

//GetHttpInfoPort return the port support info via http
func (this *NetServer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//GetRelay return whether net module can relay msg
func (this *NetServer) GetRelay() bool {
	return this.base.GetRelay()
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

//GetMaxPeerBlockHeight return the most height of valid connections
func (this *NetServer) GetMaxPeerBlockHeight() uint64 {
	return this.Np.GetNeighborMostHeight()
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
func (this *NetServer) Xmit(msg types.Message) {
	this.Np.Broadcast(msg)
}

//GetMsgChan return sync or consensus channel when msgrouter need msg input
func (this *NetServer) GetMsgChan() chan *types.MsgPayload {
	return this.NetChan
}

//Tx send data buf to peer
func (this *NetServer) Send(p *peer.Peer, msg types.Message) error {
	if p != nil {
		return p.Send(msg)
	}
	log.Warn("[p2p]send to a invalid peer")
	return errors.New("[p2p]send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (this *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p == nil {
		return false
	}
	return this.Np.NodeEstablished(p.GetID())
}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) Connect(addr string) error {
	if this.IsAddrInOutConnRecord(addr) {
		log.Debugf("[p2p]Address: %s is in OutConnectionRecord,", addr)
		return nil
	}
	if this.IsOwnAddress(addr) {
		return nil
	}
	if !this.AddrValid(addr) {
		return nil
	}

	this.connectLock.Lock()
	connCount := uint(this.GetOutConnRecordLen())
	if connCount >= config.DefConfig.P2PNode.MaxConnOutBound {
		log.Warnf("[p2p]Connect: out connections(%d) reach the max limit(%d)", connCount,
			config.DefConfig.P2PNode.MaxConnOutBound)
		this.connectLock.Unlock()
		return errors.New("[p2p]connect: out connections reach the max limit")
	}
	this.connectLock.Unlock()

	if this.IsNbrPeerAddr(addr) {
		return nil
	}
	this.connectLock.Lock()
	if addOK := this.AddOutConnectingList(addr); !addOK {
		log.Debug("[p2p]node exist in connecting list", addr)
	}
	this.connectLock.Unlock()

	isTls := config.DefConfig.P2PNode.IsTLS
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Debugf("[p2p]connect %s failed:%s", addr, err.Error())
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Debugf("[p2p]connect %s failed:%s", addr, err.Error())
			return err
		}
	}

	addr = conn.RemoteAddr().String()
	log.Debugf("[p2p]peer %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network())

	this.AddOutConnRecord(addr)
	remotePeer = peer.NewPeer()
	this.AddPeerAddress(addr, remotePeer)
	remotePeer.Link.SetAddr(addr)
	remotePeer.Link.SetConn(conn)
	remotePeer.AttachChan(this.NetChan)
	go remotePeer.Link.Rx()
	remotePeer.SetState(common.HAND)

	version := msgpack.NewVersion(this, ledger.DefLedger.GetCurrentBlockHeight())
	err = remotePeer.Send(version)
	if err != nil {
		this.RemoveFromOutConnRecord(addr)
		log.Warn(err)
		return err
	}
	return nil
}

//Halt stop all net layer logic
func (this *NetServer) Halt() {
	peers := this.Np.GetNeighbors()
	for _, p := range peers {
		p.Close()
	}
	if this.listener != nil {
		this.listener.Close()
	}
}

//establishing the connection to remote peers and listening for inbound peers
func (this *NetServer) startListening() error {
	var err error

	syncPort := this.base.GetPort()

	if syncPort == 0 {
		log.Error("[p2p]sync port invalid")
		return errors.New("[p2p]sync port invalid")
	}

	err = this.startNetListening(syncPort)
	if err != nil {
		log.Error("[p2p]start sync listening fail")
		return err
	}
	return nil
}

// startNetListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startNetListening(port uint16) error {
	var err error
	this.listener, err = createListener(port)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return errors.New("[p2p]failed to create sync listener")
	}

	go this.startNetAccept(this.listener)
	log.Infof("[p2p]start listen on sync port %d", port)
	return nil
}

//startNetAccept accepts the sync connection from the inbound peer
func (this *NetServer) startNetAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Error("[p2p]error accepting ", err.Error())
			return
		}

		log.Debug("[p2p]remote sync node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())
		if !this.AddrValid(conn.RemoteAddr().String()) {
			log.Warnf("[p2p]remote %s not in reserved list, close it ", conn.RemoteAddr())
			conn.Close()
			continue
		}

		if this.IsAddrInInConnRecord(conn.RemoteAddr().String()) {
			conn.Close()
			continue
		}

		syncAddrCount := uint(this.GetInConnRecordLen())
		if syncAddrCount >= config.DefConfig.P2PNode.MaxConnInBound {
			log.Warnf("[p2p]SyncAccept: total connections(%d) reach the max limit(%d), conn closed",
				syncAddrCount, config.DefConfig.P2PNode.MaxConnInBound)
			conn.Close()
			continue
		}

		remoteIp, err := common.ParseIPAddr(conn.RemoteAddr().String())
		if err != nil {
			log.Warn("[p2p]parse ip error ", err.Error())
			conn.Close()
			continue
		}
		connNum := this.GetIpCountInInConnRecord(remoteIp)
		if connNum >= config.DefConfig.P2PNode.MaxConnInBoundForSingleIP {
			log.Warnf("[p2p]SyncAccept: connections(%d) with ip(%s) has reach the max limit(%d), "+
				"conn closed", connNum, remoteIp, config.DefConfig.P2PNode.MaxConnInBoundForSingleIP)
			conn.Close()
			continue
		}

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddInConnRecord(addr)

		this.AddPeerAddress(addr, remotePeer)

		remotePeer.Link.SetAddr(addr)
		remotePeer.Link.SetConn(conn)
		remotePeer.AttachChan(this.NetChan)
		go remotePeer.Link.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) AddOutConnectingList(addr string) (added bool) {
	this.connectingNodes.Lock()
	defer this.connectingNodes.Unlock()
	if this.connectingNodes.ConnectingAddrs.Has(addr) {
		return false
	}

	log.Trace("[p2p]add to out connecting list", addr)
	this.connectingNodes.ConnectingAddrs.Insert(addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (this *NetServer) RemoveFromConnectingList(addr string) {
	this.connectingNodes.Lock()
	defer this.connectingNodes.Unlock()
	this.connectingNodes.ConnectingAddrs.Delete(addr)
	log.Trace("[p2p]remove from out connecting list", addr)
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) GetOutConnectingListLen() (count uint) {
	this.connectingNodes.RLock()
	defer this.connectingNodes.RUnlock()
	return uint(this.connectingNodes.ConnectingAddrs.Len())
}

//check  peer from connecting list
func (this *NetServer) IsAddrFromConnecting(addr string) bool {
	this.connectingNodes.Lock()
	defer this.connectingNodes.Unlock()
	return this.connectingNodes.ConnectingAddrs.Has(addr)
}

//find exist peer from addr map
func (this *NetServer) GetPeerFromAddr(addr string) *peer.Peer {
	var p *peer.Peer
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()

	p, ok := this.PeerAddress[addr]
	if ok {
		return p
	}
	return nil
}

//IsNbrPeerAddr return result whether the address is under connecting
func (this *NetServer) IsNbrPeerAddr(addr string) bool {
	var addrNew string
	this.Np.RLock()
	defer this.Np.RUnlock()
	for _, p := range this.Np.List {
		if p.GetState() == common.HAND || p.GetState() == common.HAND_SHAKE ||
			p.GetState() == common.ESTABLISH {
			addrNew = p.Link.GetAddr()
			if strings.Compare(addrNew, addr) == 0 {
				return true
			}
		}
	}
	return false
}

//AddPeerAddress add sync addr to peer-addr map
func (this *NetServer) AddPeerAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	log.Debugf("[p2p]AddPeerAddress %s", addr)
	this.PeerAddress[addr] = p
}

//RemovePeerAddress remove sync addr from peer-addr map
func (this *NetServer) RemovePeerAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerAddress[addr]; ok {
		delete(this.PeerAddress, addr)
		log.Debugf("[p2p]delete Sync Address %s", addr)
	}
}

//GetPeerAddressCount return length of cons addr from peer-addr map
func (this *NetServer) GetPeerAddressCount() (count uint) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	return uint(len(this.PeerAddress))
}

//AddInConnRecord add in connection to inConnRecord
func (this *NetServer) AddInConnRecord(addr string) {
	this.inConnRecord.Lock()
	defer this.inConnRecord.Unlock()
	this.inConnRecord.InConnectingAddrs.Insert(addr)
	log.Debugf("[p2p]add in record  %s", addr)
}

//IsAddrInInConnRecord return result whether addr is in inConnRecordList
func (this *NetServer) IsAddrInInConnRecord(addr string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()

	return this.inConnRecord.InConnectingAddrs.Has(addr)
}

//IsIPInInConnRecord return result whether the IP is in inConnRecordList
func (this *NetServer) IsIPInInConnRecord(ip string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var ipRecord string
	for addr := range this.inConnRecord.InConnectingAddrs {
		ipRecord, _ = common.ParseIPAddr(addr)
		if 0 == strings.Compare(ipRecord, ip) {
			return true
		}
	}
	return false
}

//RemoveInConnRecord remove in connection from inConnRecordList
func (this *NetServer) RemoveFromInConnRecord(addr string) {
	this.inConnRecord.Lock()
	defer this.inConnRecord.Unlock()
	log.Debugf("[p2p]remove in record  %s", addr)
	this.inConnRecord.InConnectingAddrs.Delete(addr)
}

//GetInConnRecordLen return length of inConnRecordList
func (this *NetServer) GetInConnRecordLen() int {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	return this.inConnRecord.InConnectingAddrs.Len()
}

//GetIpCountInInConnRecord return count of in connections with single ip
func (this *NetServer) GetIpCountInInConnRecord(ip string) uint {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var count uint
	var ipRecord string
	for addr := range this.inConnRecord.InConnectingAddrs {
		ipRecord, _ = common.ParseIPAddr(addr)
		if 0 == strings.Compare(ipRecord, ip) {
			count++
		}
	}
	return count
}

//AddOutConnRecord add out connection to outConnRecord
func (this *NetServer) AddOutConnRecord(addr string) {
	this.outConnRecord.Lock()
	defer this.outConnRecord.Unlock()
	this.outConnRecord.OutConnectingAddrs.Insert(addr)
	log.Debugf("[p2p]add out record  %s", addr)
}

//IsAddrInOutConnRecord return result whether addr is in outConnRecord
func (this *NetServer) IsAddrInOutConnRecord(addr string) bool {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	return this.outConnRecord.OutConnectingAddrs.Has(addr)
}

//RemoveOutConnRecord remove out connection from outConnRecord
func (this *NetServer) RemoveFromOutConnRecord(addr string) {
	this.outConnRecord.Lock()
	defer this.outConnRecord.Unlock()
	this.outConnRecord.OutConnectingAddrs.Delete(addr)
}

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() int {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	return this.outConnRecord.OutConnectingAddrs.Len()
}

//AddrValid whether the addr could be connect or accept
func (this *NetServer) AddrValid(addr string) bool {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		for _, ip := range config.DefConfig.P2PNode.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(addr, ip) {
				log.Info("[p2p]found reserved peer :", addr)
				return true
			}
		}
		return false
	}
	return true
}

//check own network address
func (this *NetServer) IsOwnAddress(addr string) bool {
	return addr == this.OwnAddress
}

//Set own network address
func (this *NetServer) SetOwnAddress(addr string) {
	if addr != this.OwnAddress {
		log.Infof("[p2p]set own address %s", addr)
		this.OwnAddress = addr
	}
}
