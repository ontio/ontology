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
	"net"
	"strings"
	"sync"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/connect_controller"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer() p2p.P2P {
	n := &NetServer{
		NetChan: make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		base:    &peer.PeerInfo{},
		Np:      peer.NewNbrPeers(),
	}

	n.PeerAddrMap.PeerAddress = make(map[string]*peer.Peer)

	n.init(config.DefConfig)
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base     *peer.PeerInfo
	listener net.Listener
	pid      *evtActor.PID
	NetChan  chan *types.MsgPayload
	PeerAddrMap
	Np *peer.NbrPeers

	connCtrl *connect_controller.ConnectController
	dht      *dht.DHT
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (this *NetServer) init(conf *config.OntologyConfig) error {
	dtable := dht.NewDHT()
	this.dht = dtable

	service := common.SERVICE_NODE
	if conf.Consensus.EnableConsensus {
		service = common.VERIFY_NODE
	}
	httpInfo := conf.P2PNode.HttpInfoPort
	nodePort := conf.P2PNode.NodePort
	if nodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}

	this.base = peer.NewPeerInfo(dtable.GetKadKeyId().Id, common.PROTOCOL_VERSION, uint64(service), true, httpInfo,
		nodePort, 0, config.Version)

	option, err := connect_controller.ConnCtrlOptionFromConfig(conf.P2PNode)
	if err != nil {
		return err
	}
	this.connCtrl = connect_controller.NewConnectController(this.base, dtable.GetKadKeyId(), option)

	log.Infof("[p2p]init peer ID to %s", this.base.Id.ToHexString())
	this.doRefresh()

	return nil
}

//InitListen start listening on the config port
func (this *NetServer) Start() {
	this.startListening()
}

//GetVersion return self peer`s version
func (this *NetServer) GetVersion() uint32 {
	return this.base.Version
}

//GetId return peer`s id
func (this *NetServer) GetID() uint64 {
	return this.base.Id.ToUint64()
}

func (this *NetServer) GetKId() kbucket.KadId {
	return this.base.Id
}

func (this *NetServer) GetKadKeyId() *kbucket.KadKeyId {
	return this.dht.GetKadKeyId()
}

// SetHeight sets the local's height
func (this *NetServer) SetHeight(height uint64) {
	this.base.Height = height
}

func (this *NetServer) SetPID(pid *evtActor.PID) {
	this.pid = pid
}

// GetHeight return peer's heigh
func (this *NetServer) GetHeight() uint64 {
	return this.base.Height
}

//GetTime return the last contact time of self peer
func (this *NetServer) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

//GetServices return the service state of self peer
func (this *NetServer) GetServices() uint64 {
	return this.base.Services
}

//GetPort return the sync port
func (this *NetServer) GetPort() uint16 {
	return this.base.Port
}

//GetHttpInfoPort return the port support info via http
func (this *NetServer) GetHttpInfoPort() uint16 {
	return this.base.HttpInfoPort
}

//GetRelay return whether net module can relay msg
func (this *NetServer) GetRelay() bool {
	return this.base.Relay
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

//Tx sendMsg data buf to peer
func (this *NetServer) Send(p *peer.Peer, msg types.Message) error {
	if p != nil {
		return p.Send(msg)
	}
	log.Warn("[p2p]sendMsg to a invalid peer")
	return errors.New("[p2p]sendMsg to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (this *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p == nil {
		return false
	}
	return this.Np.NodeEstablished(p.GetID())
}

func (this *NetServer) removeOldPeer(kid kbucket.KadId, remoteAddr string) {
	p := this.GetPeer(kid.ToUint64())
	if p != nil {
		n, delOK := this.DelNbrNode(kid.ToUint64())
		if delOK {
			log.Infof("[p2p] peer reconnect %s, addr: %s", kid.ToHexString(), remoteAddr)
			// Close the connection and release the node source
			n.Close()
			if this.pid != nil {
				input := &common.RemovePeerID{
					ID: kid.ToUint64(),
				}
				this.pid.Tell(input)
			}
		}
	}

}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) Connect(addr string) error {
	if this.IsNbrPeerAddr(addr) {
		return nil
	}

	peerInfo, conn, err := this.connCtrl.Connect(addr)
	if err != nil {
		return err
	}
	remotePeer := createPeer(peerInfo, conn)

	kid := remotePeer.GetKId()
	remoteAddr := remotePeer.GetAddr()
	// Obsolete node
	netServer := this
	this.removeOldPeer(kid, remoteAddr)

	if !netServer.UpdateDHT(kid) {
		return fmt.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kid.ToHexString())
	}

	remotePeer.AttachChan(netServer.NetChan)
	netServer.AddPeerAddress(remoteAddr, remotePeer)
	netServer.AddNbrNode(remotePeer)
	go remotePeer.Link.Rx()

	if netServer.pid != nil {
		input := &common.AppendPeerID{
			ID: kid.ToUint64(),
		}
		netServer.pid.Tell(input)
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

	syncPort := this.base.Port

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
	this.listener, err = connect_controller.NewListener(port, config.DefConfig.P2PNode)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return errors.New("[p2p]failed to create sync listener")
	}

	go this.startNetAccept(this.listener)
	log.Infof("[p2p]start listen on sync port %d", port)
	return nil
}

func (this *NetServer) handleClientConnection(conn net.Conn) error {
	peerInfo, conn, err := this.connCtrl.AcceptConnect(conn)
	if err != nil {
		log.Error("[p2p] allow reserved peer connection only ")
		return err
	}
	remotePeer := createPeer(peerInfo, conn)

	// Obsolete node
	kid := remotePeer.GetKId()
	this.removeOldPeer(kid, conn.RemoteAddr().String())

	this.dht.Update(kid)

	remotePeer.AttachChan(this.NetChan)
	addr := conn.RemoteAddr().String()
	this.AddNbrNode(remotePeer)
	this.AddPeerAddress(addr, remotePeer)

	go remotePeer.Link.Rx()
	if this.pid != nil {
		input := &common.AppendPeerID{
			ID: kid.ToUint64(),
		}
		this.pid.Tell(input)
	}

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

		if err := this.handleClientConnection(conn); err != nil {
			log.Errorf("[p2p] handleClientConnection error: %s", err)
			_ = conn.Close()
		}
	}
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
		if p.GetState() == common.ESTABLISH {
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

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() uint {
	return this.connCtrl.OutboundsCount()
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
	return addr == this.connCtrl.OwnAddress()
}

func (ns *NetServer) UpdateDHT(id kbucket.KadId) bool {
	ns.dht.Update(id)
	return true
}

func (ns *NetServer) RemoveDHT(id kbucket.KadId) bool {
	ns.dht.Remove(id)
	return true
}

func (ns *NetServer) BetterPeers(id kbucket.KadId, count int) []kbucket.KadId {
	return ns.dht.BetterPeers(id, count)
}

func (ns *NetServer) GetPeerStringAddr() map[uint64]string {
	return ns.Np.GetPeerStringAddr()
}

func (ns *NetServer) findSelf() {
	tick := time.NewTicker(ns.dht.RtRefreshPeriod)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			log.Debug("[dht] start to find myself")
			closer := ns.dht.BetterPeers(ns.dht.GetKadKeyId().Id, dht.AlphaValue)
			for _, id := range closer {
				log.Debugf("[dht] find closr peer %x", id)

				if id.IsPseudoKadId() {
					ns.Send(ns.GetPeer(id.ToUint64()), msgpack.NewAddrReq())
				} else {
					ns.Send(ns.GetPeer(id.ToUint64()), msgpack.NewFindNodeReq(id))
				}
			}
		}
	}
}

func (ns *NetServer) refreshCPL() {
	tick := time.NewTicker(ns.dht.RtRefreshPeriod)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			for curCPL := range ns.dht.RouteTable().Buckets {
				log.Debugf("[dht] start to refresh bucket: %d", curCPL)
				randPeer := ns.dht.RouteTable().GenRandKadId(uint(curCPL))
				closer := ns.dht.BetterPeers(randPeer, dht.AlphaValue)
				for _, pid := range closer {
					log.Debugf("[dht] find closr peer %d", pid)
					if pid.IsPseudoKadId() {
						ns.Send(ns.GetPeer(pid.ToUint64()), msgpack.NewAddrReq())
					} else {
						ns.Send(ns.GetPeer(pid.ToUint64()), msgpack.NewFindNodeReq(pid))
					}
				}
			}
		}
	}
}

func (ns *NetServer) doRefresh() {
	go ns.findSelf()
	go ns.refreshCPL()
}

func createPeer(info *peer.PeerInfo, conn net.Conn) *peer.Peer {
	remotePeer := peer.NewPeer()
	remotePeer.SetInfo(info)
	remotePeer.SetState(common.ESTABLISH)
	remotePeer.Link.UpdateRXTime(time.Now())
	remotePeer.Link.SetPort(info.Port)
	remotePeer.Link.SetAddr(conn.RemoteAddr().String())
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(info.Id.ToUint64())

	return remotePeer
}
