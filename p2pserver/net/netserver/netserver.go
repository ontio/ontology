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

	tsp "github.com/ontio/ontology/p2pserver/net/transport"
	tspCreator "github.com/ontio/ontology/p2pserver/net/transport/creator"
)

//NewNetServer return the net object in p2p
func NewNetServer() p2p.P2P {
	n := &NetServer{
		SyncChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		ConsChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
	}

	n.PeerAddrMap.PeerSyncAddress = make(map[string]*peer.Peer)
	n.PeerAddrMap.PeerConsAddress = make(map[string]*peer.Peer)

	if common.TransportType(config.DefConfig.P2PNode.TransportType) == common.LegacyTSPType {
		n.synclistener = make(map[common.TransportType]tsp.Listener, 1)
		n.conslistener = make(map[common.TransportType]tsp.Listener, 1)
	}else {
		n.synclistener = make(map[common.TransportType]tsp.Listener, 2)
		n.conslistener = make(map[common.TransportType]tsp.Listener, 2)
	}

	n.init()
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base                  peer.PeerCom
	synclistener          map[common.TransportType]tsp.Listener
	conslistener          map[common.TransportType]tsp.Listener
	SyncChan     chan *types.MsgPayload
	ConsChan     chan *types.MsgPayload
	ConnectingNodes
	PeerAddrMap
	Np                    *peer.NbrPeers
	connectLock           sync.Mutex
	inConnRecord          InConnectionRecord
	outConnRecord         OutConnectionRecord
	OwnAddress            string //network`s own address(ip : sync port),which get from version check
}

//InConnectionRecord include all addr connected
type InConnectionRecord struct {
	sync.RWMutex
	InConnectingAddrs []string
}

//OutConnectionRecord include all addr accepted
type OutConnectionRecord struct {
	sync.RWMutex
	OutConnectingAddrs []string
}

//ConnectingNodes include all addr in connecting state
type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerSyncAddress          map[string]*peer.Peer
	PeerConsAddress          map[string]*peer.Peer
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

	this.base.SetSyncPort(uint16(config.DefConfig.P2PNode.NodePort))

	if config.DefConfig.P2PNode.DualPortSupport {
		if config.DefConfig.P2PNode.NodeConsensusPort == 0 {
			log.Error("[p2p]consensus port invalid")
			return errors.New("[p2p]invalid consensus port")
		}

		this.base.SetConsPort(uint16(config.DefConfig.P2PNode.NodeConsensusPort))
	} else {
		this.base.SetConsPort(0)
	}

	this.base.SetRelay(true)

	rand.Seed(time.Now().UnixNano())
	id := rand.Uint64()

	this.base.SetID(id)

	log.Infof("[p2p]init peer ID to %d", this.base.GetID())
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
	log.Warn("[p2p]send to a invalid peer")
	return errors.New("[p2p]send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (this *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return this.Np.NodeEstablished(p.GetID())
	}
	return false
}

func (this *NetServer) Connect(addr string, isConsensus bool) error {
	tspType := common.TransportType(config.DefConfig.P2PNode.TransportType	)
	err := this.connectSub(addr, isConsensus, tspType)
	switch err.(type){
	case *tsp.DialError:
		if tspType != common.LegacyTSPType {
			log.Tracef("[p2p]Connect to %s dial err by transport %s and switch to transport %s",
				addr,
				common.GetTransportTypeString(tspType),
				common.GetTransportTypeString(common.LegacyTSPType))
			return this.tryLegacyConnect(addr, isConsensus)
		}else {
			log.Errorf("[p2p]DialError by transport %s", common.GetTransportTypeString(tspType))
			return err
		}
	default:
		return err
	}

	return err
}

func (this *NetServer) tryLegacyConnect(addr string, isConsensus bool) error {
	log.Tracef("[p2p]tryLegacyConnect to %s dial err by transport %s",
		addr,
		common.GetTransportTypeString(common.LegacyTSPType))
	err := this.connectSub(addr, isConsensus, common.LegacyTSPType)
	switch err.(type) {
	case *tsp.DialError:
		log.Tracef("[p2p]tryLegacyConnect to %s dial err by transport %s",
			addr,
			common.GetTransportTypeString(common.LegacyTSPType))
		return this.connectSub(addr, isConsensus, common.LegacyTSPType)

	default:
		return err
	}

	return err
}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) connectSub(addr string, isConsensus bool, tspType common.TransportType) error {
	if this.IsAddrInOutConnRecord(addr) {
		log.Debugf("[p2p]Address: %s Consensus: %v is in OutConnectionRecord,", addr, isConsensus)
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

	if this.IsNbrPeerAddr(addr, isConsensus) {
		return nil
	}
	this.connectLock.Lock()
	if added := this.AddOutConnectingList(addr); added == false {
		log.Debug("[p2p]node exist in connecting list", addr)
	}
	this.connectLock.Unlock()

	transport, err := tspCreator.GetTransportFactory().GetTransport(tspType)
	if err != nil {
		log.Errorf("[p2p]Get the transport of %s, err:%s", common.GetTransportTypeString(tspType), err.Error())
		return err
	}

	conn, err := transport.Dial(addr)
	if err != nil {
		this.RemoveFromConnectingList(addr)
		log.Debugf("[p2p]connect %s failed:%s by transport %s", addr, err.Error(), common.GetTransportTypeString(tspType))
		return err
	}

	addr = conn.RemoteAddr().String()
	log.Debugf("[p2p]peer %s connect with %s with %s by transport %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network(), common.GetTransportTypeString(tspType))

	var remotePeer *peer.Peer
	if !isConsensus {
		this.AddOutConnRecord(addr)
		remotePeer = peer.NewPeer()
		this.AddPeerSyncAddress(addr, remotePeer)
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(this.SyncChan)
		go remotePeer.SyncLink.Rx()
		remotePeer.SetSyncState(common.HAND)
		remotePeer.SetTransportType(tspType)

	} else {
		remotePeer = peer.NewPeer() //would merge with a exist peer in versionhandle
		this.AddPeerConsAddress(addr, remotePeer)
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetConsState(common.HAND)
		remotePeer.SetTransportType(tspType)
	}
	version := msgpack.NewVersion(this, isConsensus, ledger.DefLedger.GetCurrentBlockHeight())
	err = remotePeer.Send(version, isConsensus)
	if err != nil {
		if !isConsensus {
			this.RemoveFromOutConnRecord(addr)
		}
		log.Warn(err)
		return err
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

	for _, syncL := range this.synclistener {
		if syncL != nil {
			syncL.Close()
		}
	}
	for _, consL := range this.conslistener {
		if consL != nil {
			consL.Close()
		}
	}
}

//establishing the connection to remote peers and listening for inbound peers
func (this *NetServer) startListening() error {
	var err error

	syncPort := this.base.GetSyncPort()
	consPort := this.base.GetConsPort()
	syncPortLegacy := this.base.GetSyncPort()
	consPortLegacy := this.base.GetConsPort()

	if syncPort == 0  || syncPortLegacy == 0{
		log.Error("[p2p]sync port invalid")
		return errors.New("[p2p]sync port invalid")
	}

	tspType := common.TransportType(config.DefConfig.P2PNode.TransportType)
	err = this.startSyncListening(uint16(syncPort), tspType)
	if err != nil {
		log.Error("[p2p]start sync TCP listening fail")
	}

	if tspType != common.LegacyTSPType {
		err = this.startSyncListening(uint16(syncPortLegacy), common.LegacyTSPType)
		if err != nil {
			log.Errorf("[p2p]start sync listening fail by %s", common.GetTransportTypeString(common.LegacyTSPType))
			return err
		}
	}

	//consensus
	if config.DefConfig.P2PNode.DualPortSupport == false {
		log.Debug("[p2p]dual port mode not supported,keep single link")
		return nil
	}
	if consPort == 0 || consPort == syncPort {
		//still work
		log.Warn("[p2p]consensus port invalid,keep single link")
	} else {
		err = this.startConsListening(uint16(consPort), tspType)
		if err != nil {
			log.Errorf("[p2p]start consensus %s listening fail", common.GetTransportTypeString(tspType))
		}

		if tspType != common.LegacyTSPType {
			err = this.startConsListening(uint16(consPortLegacy), common.LegacyTSPType)
			if err != nil {
				log.Errorf("[p2p]start consensus %s listening fail", common.GetTransportTypeString(common.LegacyTSPType))
				return err
			}
		}
	}
	return nil
}

// startSyncListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startSyncListening(port uint16, tspType common.TransportType) error {
	transport, err := tspCreator.GetTransportFactory().GetTransport(tspType)
	if err != nil {
		log.Errorf("[p2p]Get the transport of %s, err:%s", common.GetTransportTypeString(tspType), err.Error())
		return err
	}

	listener, err := transport.Listen(port)
	if err != nil {
		errStr := fmt.Sprintf("[p2p]failed to create sync %s listener", common.GetTransportTypeString(tspType))
		log.Error(errStr)
		return errors.New(errStr)
	}

	this.synclistener[tspType] = listener

	go this.startSyncAccept(listener, tspType)
	log.Tracef("[p2p]start listen on sync %s port %d", common.GetTransportTypeString(tspType), port)
	return nil
}

// startConsListening starts a consensus listener on the port for the inbound peer
func (this *NetServer) startConsListening(port uint16, tspType common.TransportType) error {
	transport, err := tspCreator.GetTransportFactory().GetTransport(tspType)
	if err != nil {
		log.Errorf("[p2p]Get the transport of %s, err:%s", common.GetTransportTypeString(tspType), err.Error())
		return err
	}

	listener, err := transport.Listen(port)
	if err != nil {
		errStr := fmt.Sprintf("[p2p]failed to create cons %s listener", common.GetTransportTypeString(tspType))
		log.Error(errStr)
		return errors.New(errStr)
	}

	this.conslistener[tspType] = listener

	go this.startConsAccept(listener, tspType)
	log.Tracef("[p2p]Start listen on consensus %s port %d", common.GetTransportTypeString(tspType), port)
	return nil
}

//startSyncAccept accepts the sync connection from the inbound peer
func (this *NetServer) startSyncAccept(listener tsp.Listener, tspType common.TransportType) {
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
		this.AddPeerSyncAddress(addr, remotePeer)
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(this.SyncChan)
		go remotePeer.SyncLink.Rx()
		remotePeer.SetTransportType(tspType)
	}
}

//startConsAccept accepts the consensus connnection from the inbound peer
func (this *NetServer) startConsAccept(listener tsp.Listener, tspType common.TransportType) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("[p2p]error accepting ", err.Error())
			return
		}
		log.Debug("[p2p]remote cons node connect with ",
			conn.RemoteAddr(), conn.LocalAddr())
		if !this.AddrValid(conn.RemoteAddr().String()) {
			log.Warnf("[p2p]remote %s not in reserved list, close it ", conn.RemoteAddr())
			conn.Close()
			continue
		}

		remoteIp, err := common.ParseIPAddr(conn.RemoteAddr().String())
		if err != nil {
			log.Warn("[p2p]parse ip error ", err.Error())
			conn.Close()
			continue
		}
		if !this.IsIPInInConnRecord(remoteIp) {
			conn.Close()
			continue
		}

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddPeerConsAddress(addr, remotePeer)

		remotePeer.ConsLink.SetAddr(addr)
		log.Tracef("[p2p]Set remote peer conslink conn during startConsAccept, remoteConsaddr =%s, tspType=%s", addr, common.GetTransportTypeString(tspType))
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(this.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetTransportType(tspType)
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) AddOutConnectingList(addr string) (added bool) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	for _, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	log.Trace("[p2p]add to out connecting list", addr)
	this.ConnectingAddrs = append(this.ConnectingAddrs, addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (this *NetServer) RemoveFromConnectingList(addr string) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	addrs := this.ConnectingAddrs[:0]
	for _, a := range this.ConnectingAddrs {
		if a != addr {
			addrs = append(addrs, a)
		}
	}
	log.Trace("[p2p]remove from out connecting list", addr)
	this.ConnectingAddrs = addrs
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) GetOutConnectingListLen() (count uint) {
	this.ConnectingNodes.RLock()
	defer this.ConnectingNodes.RUnlock()
	return uint(len(this.ConnectingAddrs))
}

//check  peer from connecting list
func (this *NetServer) IsAddrFromConnecting(addr string) bool {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	for _, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
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
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	log.Debugf("[p2p]AddPeerSyncAddress %s", addr)
	this.PeerSyncAddress[addr] = p
}

//AddPeerConsAddress add cons addr to peer-addr map
func (this *NetServer) AddPeerConsAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	log.Debugf("[p2p]AddPeerConsAddress %s", addr)
	this.PeerConsAddress[addr] = p
}

//RemovePeerSyncAddress remove sync addr from peer-addr map
func (this *NetServer) RemovePeerSyncAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerSyncAddress[addr]; ok {
		delete(this.PeerSyncAddress, addr)
		log.Debugf("[p2p]delete Sync Address %s", addr)
	}
}

//RemovePeerConsAddress remove cons addr from peer-addr map
func (this *NetServer) RemovePeerConsAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerConsAddress[addr]; ok {
		delete(this.PeerConsAddress, addr)
		log.Debugf("[p2p]delete Cons Address %s", addr)
	}
}

//GetPeerSyncAddressCount return length of cons addr from peer-addr map
func (this *NetServer) GetPeerSyncAddressCount() (count uint) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	return uint(len(this.PeerSyncAddress))
}

//AddInConnRecord add in connection to inConnRecord
func (this *NetServer) AddInConnRecord(addr string) {
	this.inConnRecord.Lock()
	defer this.inConnRecord.Unlock()
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return
		}
	}
	this.inConnRecord.InConnectingAddrs = append(this.inConnRecord.InConnectingAddrs, addr)
	log.Debugf("[p2p]add in record  %s", addr)
}

//IsAddrInInConnRecord return result whether addr is in inConnRecordList
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

//IsIPInInConnRecord return result whether the IP is in inConnRecordList
func (this *NetServer) IsIPInInConnRecord(ip string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var ipRecord string
	for _, addr := range this.inConnRecord.InConnectingAddrs {
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
	addrs := []string{}
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) != 0 {
			addrs = append(addrs, a)
		}
	}
	log.Debugf("[p2p]remove in record  %s", addr)
	this.inConnRecord.InConnectingAddrs = addrs
}

//GetInConnRecordLen return length of inConnRecordList
func (this *NetServer) GetInConnRecordLen() int {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	return len(this.inConnRecord.InConnectingAddrs)
}

//GetIpCountInInConnRecord return count of in connections with single ip
func (this *NetServer) GetIpCountInInConnRecord(ip string) uint {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var count uint
	var ipRecord string
	for _, addr := range this.inConnRecord.InConnectingAddrs {
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
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return
		}
	}
	this.outConnRecord.OutConnectingAddrs = append(this.outConnRecord.OutConnectingAddrs, addr)
	log.Debugf("[p2p]add out record  %s", addr)
}

//IsAddrInOutConnRecord return result whether addr is in outConnRecord
func (this *NetServer) IsAddrInOutConnRecord(addr string) bool {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
}

//RemoveOutConnRecord remove out connection from outConnRecord
func (this *NetServer) RemoveFromOutConnRecord(addr string) {
	this.outConnRecord.Lock()
	defer this.outConnRecord.Unlock()
	addrs := []string{}
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) != 0 {
			addrs = append(addrs, a)
		}
	}
	log.Debugf("[p2p]remove out record  %s", addr)
	this.outConnRecord.OutConnectingAddrs = addrs
}

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() int {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	return len(this.outConnRecord.OutConnectingAddrs)
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
	if addr == this.OwnAddress {
		return true
	}
	return false
}

//Set own network address
func (this *NetServer) SetOwnAddress(addr string) {
	if addr != this.OwnAddress {
		log.Infof("[p2p]set own address %s", addr)
		this.OwnAddress = addr
	}
}
