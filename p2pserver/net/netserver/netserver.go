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
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer() p2p.P2P {
	n := &NetServer{
		SyncChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		ConsChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
	}

	n.PeerAddrMap.PeerSyncAddress = make(map[string]*peer.Peer)
	n.PeerAddrMap.PeerConsAddress = make(map[string]*peer.Peer)
	n.outConnRecord.OutConnectingAddrs = make(map[string]int)

	n.init()
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base         peer.PeerCom
	synclistener net.Listener
	conslistener net.Listener
	SyncChan     chan *types.MsgPayload
	ConsChan     chan *types.MsgPayload
	ConnectingNodes
	PeerAddrMap
	Np            *peer.NbrPeers
	connectLock   sync.Mutex
	inConnRecord  InConnectionRecord
	outConnRecord OutConnectionRecord
}

//InConnectionRecord include all addr connected
type InConnectionRecord struct {
	sync.RWMutex
	InConnectingAddrs []string
}

//OutConnectionRecord include all addr accepted
type OutConnectionRecord struct {
	sync.RWMutex
	OutConnectingAddrs map[string]int
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
func (this *NetServer) init() error {
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

	rand.Seed(time.Now().UnixNano())
	id := rand.Uint64()

	this.base.SetID(id)

	log.Infof("init peer ID to %d", this.base.GetID())
	this.Np = &peer.NbrPeers{}
	this.Np.Init()

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
func (this *NetServer) Xmit(msg types.Message, isCons bool) {
	this.Np.Broadcast(msg, isCons)
}

//GetMsgChan return sync or consensus channel when msgrouter need msg input
func (this *NetServer) GetMsgChan(isConsensus bool) chan *types.MsgPayload {
	if isConsensus {
		return this.ConsChan
	} else {
		return this.SyncChan
	}
}

//Tx send data buf to peer
func (this *NetServer) Send(p *peer.Peer, msg types.Message, isConsensus bool) error {
	if p != nil {
		if config.DefConfig.P2PNode.DualPortSupport == false {
			return p.Send(msg, false)
		}
		return p.Send(msg, isConsensus)
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
	if !this.AddrValid(addr) {
		return nil
	}
	this.connectLock.Lock()
	defer this.connectLock.Unlock()
	connCount := uint(this.GetOutConnRecordLen())
	if connCount >= config.DefConfig.P2PNode.MaxConnOutBound {
		log.Warnf("Connect: out connections(%d) reach the max limit(%d)", connCount,
			config.DefConfig.P2PNode.MaxConnOutBound)
		this.PrintOutConnRecord()
		return errors.New("connect: out connections reach the max limit")
	}
	if this.IsNbrPeerAddr(addr, isConsensus) {
		return nil
	}
	if added := this.AddOutConnectingList(addr); added == false {
		p := this.GetPeerFromAddr(addr)
		if p != nil {
			if p.SyncLink.Valid() {
				log.Info("node exist in connecting list", addr)
				return errors.New("node exist in connecting list")
			}
		}
		this.RemoveFromConnectingList(addr)
	}
	this.AddOutConnRecord(addr, common.HAND)

	isTls := config.DefConfig.P2PNode.IsTLS
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			this.RemoveFromOutConnRecord(addr)
			this.RemoveFromConnectingList(addr)
			log.Error("connect failed: ", err)
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			this.RemoveFromOutConnRecord(addr)
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
		version := msgpack.NewVersion(this, false, ledger.DefLedger.GetCurrentBlockHeight())
		err := remotePeer.SyncLink.Tx(version)
		if err != nil {
			this.RemoveFromOutConnRecord(addr)
			log.Error(err)
			return err
		}
	} else {
		remotePeer = peer.NewPeer() //would merge with a exist peer in versionhandle
		this.AddPeerConsAddress(addr, remotePeer)
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetConsState(common.HAND)
		version := msgpack.NewVersion(this, true, ledger.DefLedger.GetCurrentBlockHeight())
		err := remotePeer.ConsLink.Tx(version)
		if err != nil {
			log.Error(err)
			return err
		}
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
		log.Error("sync port invalid")
		return errors.New("sync port invalid")
	}

	err = this.startSyncListening(syncPort)
	if err != nil {
		log.Error("start sync listening fail")
		return err
	}

	//consensus
	if config.DefConfig.P2PNode.DualPortSupport == false {
		log.Info("dual port mode not supported,keep single link")
		return nil
	}
	if consPort == 0 || consPort == syncPort {
		//still work
		log.Error("consensus port invalid,keep single link")
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
	log.Infof("start listen on sync port %d", port)
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
			log.Error("error accepting ", err.Error())
			continue
		}
		if !this.AddrValid(conn.RemoteAddr().String()) {
			log.Infof("remote %s not in reserved list close it ", conn.RemoteAddr())
			conn.Close()
		}
		log.Info("remote sync node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())

		if this.IsAddrInInConnRecord(conn.RemoteAddr().String()) {
			return
		}

		syncAddrCount := uint(this.GetInConnRecordLen())
		if syncAddrCount >= config.DefConfig.P2PNode.MaxConnInBound {
			log.Warnf("SyncAccept: total connections(%d) reach the max limit(%d), conn closed",
				syncAddrCount, config.DefConfig.P2PNode.MaxConnInBound)
			conn.Close()
			this.PrintInConnRecord()
			continue
		}
/*
		remoteAddr := conn.RemoteAddr().String()
		colonPos := strings.LastIndex(remoteAddr, ":")
		if colonPos == -1 {
			colonPos = len(remoteAddr)
		}
		remoteIp := remoteAddr[:colonPos]
		connNum := this.GetInConnCountWithSingleIp(remoteIp)
		if connNum >= config.DefConfig.P2PNode.MaxConnInBoundForSingleIP {
			log.Warnf("SyncAccept: connections(%d) with ip(%s) has reach the max limit(%d), "+
				"conn closed", connNum, remoteIp, config.DefConfig.P2PNode.MaxConnInBoundForSingleIP)
			conn.Close()
			continue
		}
*/
		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddInConnRecord(addr)

		this.AddPeerSyncAddress(addr, remotePeer)

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
			log.Error("error accepting ", err.Error())
			continue
		}
		if !this.AddrValid(conn.RemoteAddr().String()) {
			log.Infof("remote %s not in reserved list close it ", conn.RemoteAddr())
			conn.Close()
		}
		log.Info("remote cons node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddPeerConsAddress(addr, remotePeer)

		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) AddOutConnectingList(addr string) (added bool) {
	this.ConnectingNodes.RLock()
	defer this.ConnectingNodes.RUnlock()
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
	this.ConnectingNodes.RLock()
	defer this.ConnectingNodes.RUnlock()
	addrs := []string{}
	for i, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(this.ConnectingAddrs[:i], this.ConnectingAddrs[i+1:]...)
		}
	}
	this.ConnectingAddrs = addrs
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) GetOutConnectingListLen() (count uint) {
	this.ConnectingNodes.RLock()
	defer this.ConnectingNodes.RUnlock()
	return uint(len(this.ConnectingAddrs))
}

//find exist peer from addr map
func (this *NetServer) GetPeerFromAddr(addr string) *peer.Peer {
	var p *peer.Peer
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()

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
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	this.PeerSyncAddress[addr] = p
}

//AddPeerConsAddress add cons addr to peer-addr map
func (this *NetServer) AddPeerConsAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	this.PeerConsAddress[addr] = p
}

//RemovePeerSyncAddress remove sync addr from peer-addr map
func (this *NetServer) RemovePeerSyncAddress(addr string) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	if _, ok := this.PeerSyncAddress[addr]; ok {
		delete(this.PeerSyncAddress, addr)
	}
}

//RemovePeerConsAddress remove cons addr from peer-addr map
func (this *NetServer) RemovePeerConsAddress(addr string) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	if _, ok := this.PeerConsAddress[addr]; ok {
		delete(this.PeerConsAddress, addr)
	}
}

//GetPeerSyncAddressCount return length of cons addr from peer-addr map
func (this *NetServer) GetPeerSyncAddressCount()(count uint) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	return uint(len(this.PeerSyncAddress))
}

//--------------------------------------------------------------------------

//AddInConnRecord add in connection to inConnRecord
func (this *NetServer) AddInConnRecord(addr string) {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return
		}
	}
	this.inConnRecord.InConnectingAddrs = append(this.inConnRecord.InConnectingAddrs, addr)
}

//IsAddrInInConnRecord return result whether addr is in inConnRecord
func (this *NetServer) IsAddrInInConnRecord(addr string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
}

//RemoveInConnRecord remove in connection from inConnRecord
func (this *NetServer) RemoveFromInConnRecord(addr string) {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	addrs := []string{}
	for i, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(this.inConnRecord.InConnectingAddrs[:i],
				this.inConnRecord.InConnectingAddrs[i+1:]...)
		}
	}
	this.inConnRecord.InConnectingAddrs = addrs
}

//GetInConnRecordLen return length of inConnRecord
func (this *NetServer) GetInConnRecordLen() int {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	return len(this.inConnRecord.InConnectingAddrs)
}

func (this *NetServer) PrintInConnRecord() {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	log.Warn("---------------PrintInConnRecord---Bgn---------------")
	for _, addr := range this.inConnRecord.InConnectingAddrs {
		log.Warn(addr)
	}
	log.Warn("---------------PrintInConnRecord---End---------------")
}
//--------------------------------------------------------------------------

//AddOutConnRecord add out connection to outConnRecord
func (this *NetServer) AddOutConnRecord(addr string, status int) {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	if _, ok := this.outConnRecord.OutConnectingAddrs[addr]; !ok {
		this.outConnRecord.OutConnectingAddrs[addr] = status
	}
}

//IsAddrInOutConnRecord return result whether addr is in outConnRecord
func (this *NetServer) IsAddrInOutConnRecord(addr string) bool {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	_, ok := this.outConnRecord.OutConnectingAddrs[addr]
	return ok
}

//RemoveOutConnRecord remove out connection from outConnRecord
func (this *NetServer) RemoveFromOutConnRecord(addr string) {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	if _, ok := this.outConnRecord.OutConnectingAddrs[addr]; ok {
		delete(this.outConnRecord.OutConnectingAddrs, addr)
	}
}

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() int {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	return len(this.outConnRecord.OutConnectingAddrs)
}

func (this *NetServer) PrintOutConnRecord() {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	log.Warn("---------------PrintOutConnRecord---Bgn---------------")
	for k, v := range this.outConnRecord.OutConnectingAddrs {
		log.Warn(k, v)
	}
	log.Warn("---------------PrintOutConnRecord---End---------------")
}

//--------------------------------------------------------------------------

//GetInConnCountWithSingleIp return count of cons with single ip
func (this *NetServer) GetInConnCountWithSingleIp(ip string) uint {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var count uint
	for _, addr := range this.inConnRecord.InConnectingAddrs {
		if strings.Contains(addr, ip) {
			count++
		}
	}
	return count
}

//AddrValid whether the addr could be connect or accept
func (this *NetServer) AddrValid(addr string) bool {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedPeers) > 0 {
		for _, ip := range config.DefConfig.P2PNode.ReservedPeers {
			if strings.HasPrefix(addr, ip) {
				log.Info("found reserved peer :", addr)
				return true
			}
		}
		return false
	}
	return true

}
