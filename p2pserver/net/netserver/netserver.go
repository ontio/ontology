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

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/connect_controller"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols"
)

//NewNetServer return the net object in p2p
func NewNetServer(protocol protocols.Protocol, conf *config.P2PNodeConfig) (p2p.P2P, error) {
	n := &NetServer{
		NetChan:  make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		base:     &peer.PeerInfo{},
		Np:       peer.NewNbrPeers(),
		protocol: protocol,
	}

	n.PeerAddrMap.PeerAddress = make(map[string]*peer.Peer)

	err := n.init(conf)
	if err != nil {
		return nil, err
	}
	return n, nil
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base     *peer.PeerInfo
	listener net.Listener
	protocol protocols.Protocol
	NetChan  chan *types.MsgPayload
	PeerAddrMap
	Np *peer.NbrPeers

	connCtrl *connect_controller.ConnectController
	dht      *dht.DHT

	stopRecvCh chan bool // To stop sync channel
}

// processMessage loops to handle the message from the network
func (this *NetServer) processMessage(channel chan *types.MsgPayload,
	stopCh chan bool) {
	for {
		select {
		case data, ok := <-channel:
			if ok {
				sender := this.GetPeer(data.Id)
				if sender == nil {
					log.Warnf("[router] remote peer %d invalid.", data.Id)
					continue
				}

				ctx := protocols.NewContext(sender, this, data.PayloadSize)
				go this.protocol.HandlePeerMessage(ctx, data.Payload)
			}
		case <-stopCh:
			return
		}
	}
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (this *NetServer) init(conf *config.P2PNodeConfig) error {
	dtable := dht.NewDHT()
	this.dht = dtable

	httpInfo := conf.HttpInfoPort
	nodePort := conf.NodePort
	if nodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}

	this.base = peer.NewPeerInfo(dtable.GetPeerKeyId().Id, common.PROTOCOL_VERSION, common.SERVICE_NODE, true, httpInfo,
		nodePort, 0, config.Version, "")

	option, err := connect_controller.ConnCtrlOptionFromConfig(conf)
	if err != nil {
		return err
	}
	this.connCtrl = connect_controller.NewConnectController(this.base, dtable.GetPeerKeyId(), option)

	log.Infof("[p2p]init peer ID to %s", this.base.Id.ToHexString())

	return nil
}

//InitListen start listening on the config port
func (this *NetServer) Start() {
	this.protocol.HandleSystemMessage(this, protocols.NetworkStart{})
	this.startListening()
	go this.processMessage(this.NetChan, this.stopRecvCh)

	this.doRefresh()

	log.Debug("[p2p]MessageRouter start to parse p2p message...")
}

//GetVersion return self peer`s version
func (this *NetServer) GetHostInfo() *peer.PeerInfo {
	return this.base
}

//GetId return peer`s id
func (this *NetServer) GetID() common.PeerId {
	return this.base.Id
}

func (this *NetServer) GetPeerKeyId() *common.PeerKeyId {
	return this.dht.GetPeerKeyId()
}

// SetHeight sets the local's height
func (this *NetServer) SetHeight(height uint64) {
	this.base.Height = height
}

// GetHeight return peer's heigh
func (this *NetServer) GetHeight() uint64 {
	return this.base.Height
}

// GetPeer returns a peer with the peer id
func (this *NetServer) GetPeer(id common.PeerId) *peer.Peer {
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
func (this *NetServer) DelNbrNode(id common.PeerId) (*peer.Peer, bool) {
	return this.Np.DelNbrNode(id)
}

//GetNeighbors return all nbr peer
func (this *NetServer) GetNeighbors() []*peer.Peer {
	return this.Np.GetNeighbors()
}

//NodeEstablished return whether a peer is establish with self according to id
func (this *NetServer) NodeEstablished(id common.PeerId) bool {
	return this.Np.NodeEstablished(id)
}

//Xmit called by actor, broadcast msg
func (this *NetServer) Xmit(msg types.Message) {
	this.Np.Broadcast(msg)
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

func (this *NetServer) removeOldPeer(kid common.PeerId, remoteAddr string) {
	p := this.GetPeer(kid)
	if p != nil {
		n, delOK := this.DelNbrNode(kid)
		if delOK {
			log.Infof("[p2p] peer reconnect %s, addr: %s", kid.ToHexString(), remoteAddr)
			// Close the connection and release the node source
			n.Close()

			this.protocol.HandleSystemMessage(this, protocols.PeerDisConnected{Info: n.Info})
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

	kid := remotePeer.GetID()
	remoteAddr := remotePeer.GetAddr()
	// Obsolete node
	this.removeOldPeer(kid, remoteAddr)

	if !this.UpdateDHT(kid) {
		return fmt.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kid.ToHexString())
	}

	remotePeer.AttachChan(this.NetChan)
	this.AddPeerAddress(remoteAddr, remotePeer)
	this.AddNbrNode(remotePeer)
	go remotePeer.Link.Rx()

	this.protocol.HandleSystemMessage(this, protocols.PeerConnected{Info: remotePeer.Info})
	return nil
}

func (this *NetServer) notifyPeerConnected(p *peer.PeerInfo) {
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

	if this.stopRecvCh != nil {
		this.stopRecvCh <- true
	}
	this.protocol.HandleSystemMessage(this, protocols.NetworkStop{})
}

//establishing the connection to remote peers and listening for inbound peers
func (this *NetServer) startListening() error {
	syncPort := this.base.Port
	if syncPort == 0 {
		log.Error("[p2p]sync port invalid")
		return errors.New("[p2p]sync port invalid")
	}

	err := this.startNetListening(syncPort, config.DefConfig.P2PNode)
	if err != nil {
		log.Error("[p2p]start sync listening fail")
		return err
	}
	return nil
}

// startNetListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startNetListening(port uint16, config *config.P2PNodeConfig) error {
	var err error
	this.listener, err = connect_controller.NewListener(port, config)
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
		return err
	}
	remotePeer := createPeer(peerInfo, conn)

	// Obsolete node
	kid := remotePeer.GetID()
	this.removeOldPeer(kid, conn.RemoteAddr().String())

	this.dht.Update(kid)

	remotePeer.AttachChan(this.NetChan)
	addr := conn.RemoteAddr().String()
	this.AddNbrNode(remotePeer)
	this.AddPeerAddress(addr, remotePeer)

	go remotePeer.Link.Rx()
	this.protocol.HandleSystemMessage(this, protocols.PeerConnected{Info: remotePeer.Info})
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
			log.Warnf("[p2p] client connect error: %s", err)
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

func (ns *NetServer) UpdateDHT(id common.PeerId) bool {
	ns.dht.Update(id)
	return true
}

func (ns *NetServer) RemoveDHT(id common.PeerId) bool {
	ns.dht.Remove(id)
	return true
}

func (ns *NetServer) BetterPeers(id common.PeerId, count int) []common.PeerId {
	return ns.dht.BetterPeers(id, count)
}

func (ns *NetServer) GetPeerStringAddr() map[common.PeerId]string {
	return ns.Np.GetPeerStringAddr()
}

func (ns *NetServer) findSelf() {
	tick := time.NewTicker(ns.dht.RtRefreshPeriod)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			log.Debug("[dht] start to find myself")
			closer := ns.dht.BetterPeers(ns.dht.GetPeerKeyId().Id, dht.AlphaValue)
			for _, id := range closer {
				log.Debugf("[dht] find closr peer %x", id)

				if id.IsPseudoPeerId() {
					ns.Send(ns.GetPeer(id), msgpack.NewAddrReq())
				} else {
					ns.Send(ns.GetPeer(id), msgpack.NewFindNodeReq(id))
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
					if pid.IsPseudoPeerId() {
						ns.Send(ns.GetPeer(pid), msgpack.NewAddrReq())
					} else {
						ns.Send(ns.GetPeer(pid), msgpack.NewFindNodeReq(randPeer))
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
	remotePeer.Link.SetAddr(conn.RemoteAddr().String())
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(info.Id)

	return remotePeer
}
