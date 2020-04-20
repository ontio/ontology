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
	"net"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/connect_controller"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer(protocol p2p.Protocol, conf *config.P2PNodeConfig) (*NetServer, error) {
	n := &NetServer{
		NetChan:    make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		base:       &peer.PeerInfo{},
		Np:         NewNbrPeers(),
		protocol:   protocol,
		stopRecvCh: make(chan bool),
	}

	err := n.init(conf)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func NewCustomNetServer(id *common.PeerKeyId, info *peer.PeerInfo, proto p2p.Protocol,
	listener net.Listener, opt connect_controller.ConnCtrlOption) *NetServer {
	n := &NetServer{
		base:       info,
		listener:   listener,
		protocol:   proto,
		NetChan:    make(chan *types.MsgPayload, common.CHAN_CAPABILITY),
		Np:         NewNbrPeers(),
		stopRecvCh: make(chan bool),
	}
	n.connCtrl = connect_controller.NewConnectController(info, id, opt)

	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base     *peer.PeerInfo
	listener net.Listener
	protocol p2p.Protocol
	NetChan  chan *types.MsgPayload
	Np       *NbrPeers

	connCtrl *connect_controller.ConnectController

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
					log.Warnf("[router] remote peer %s invalid.", data.Id.ToHexString())
					continue
				}

				ctx := p2p.NewContext(sender, this, data.PayloadSize)
				go this.protocol.HandlePeerMessage(ctx, data.Payload)
			}
		case <-stopCh:
			return
		}
	}
}

//init initializes attribute of network server
func (this *NetServer) init(conf *config.P2PNodeConfig) error {
	keyId := common.RandPeerKeyId()

	httpInfo := conf.HttpInfoPort
	nodePort := conf.NodePort
	if nodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}

	this.base = peer.NewPeerInfo(keyId.Id, common.PROTOCOL_VERSION, common.SERVICE_NODE, true, httpInfo,
		nodePort, 0, config.Version, "")

	option, err := connect_controller.ConnCtrlOptionFromConfig(conf)
	if err != nil {
		return err
	}
	this.connCtrl = connect_controller.NewConnectController(this.base, keyId, option)

	syncPort := this.base.Port
	if syncPort == 0 {
		log.Error("[p2p]sync port invalid")
		return errors.New("[p2p]sync port invalid")
	}
	this.listener, err = connect_controller.NewListener(syncPort, config.DefConfig.P2PNode)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return errors.New("[p2p]failed to create sync listener")
	}

	log.Infof("[p2p]init peer ID to %s", this.base.Id.ToHexString())

	return nil
}

//InitListen start listening on the config port
func (this *NetServer) Start() error {
	this.protocol.HandleSystemMessage(this, p2p.NetworkStart{})
	go this.startNetAccept(this.listener)
	log.Infof("[p2p]start listen on sync port %d", this.base.Port)
	go this.processMessage(this.NetChan, this.stopRecvCh)

	log.Debug("[p2p]MessageRouter start to parse p2p message...")
	return nil
}

//GetVersion return self peer`s version
func (this *NetServer) GetHostInfo() *peer.PeerInfo {
	return this.base
}

//GetId return peer`s id
func (this *NetServer) GetID() common.PeerId {
	return this.base.Id
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

func (this *NetServer) ReplacePeer(remotePeer *peer.Peer) {
	old := this.Np.ReplacePeer(remotePeer, this)
	if old != nil {
		old.Close()
	}
}

//GetNeighbors return all nbr peer
func (this *NetServer) GetNeighbors() []*peer.Peer {
	return this.Np.GetNeighbors()
}

//Broadcast called by actor, broadcast msg
func (this *NetServer) Broadcast(msg types.Message) {
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

//Connect used to connect net address under sync or cons mode
func (this *NetServer) Connect(addr string) {
	err := this.connect(addr)
	if err != nil {
		log.Debugf("%s connecting to %s failed, err: %s", this.base.Addr, addr, err)
	}
}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) connect(addr string) error {
	peerInfo, conn, err := this.connCtrl.Connect(addr)
	if err != nil {
		return err
	}
	remotePeer := createPeer(peerInfo, conn)

	remotePeer.AttachChan(this.NetChan)
	this.ReplacePeer(remotePeer)
	go remotePeer.Link.Rx()

	this.protocol.HandleSystemMessage(this, p2p.PeerConnected{Info: remotePeer.Info})
	return nil
}

func (this *NetServer) notifyPeerConnected(p *peer.PeerInfo) {
	this.protocol.HandleSystemMessage(this, p2p.PeerConnected{Info: p})
}

func (this *NetServer) notifyPeerDisconnected(p *peer.PeerInfo) {
	this.protocol.HandleSystemMessage(this, p2p.PeerDisConnected{Info: p})
}

//Stop stop all net layer logic
func (this *NetServer) Stop() {
	peers := this.Np.GetNeighbors()
	for _, p := range peers {
		p.Close()
	}

	if this.listener != nil {
		_ = this.listener.Close()
	}
	close(this.stopRecvCh)
	this.protocol.HandleSystemMessage(this, p2p.NetworkStop{})
}

func (this *NetServer) handleClientConnection(conn net.Conn) error {
	peerInfo, conn, err := this.connCtrl.AcceptConnect(conn)
	if err != nil {
		return err
	}
	remotePeer := createPeer(peerInfo, conn)
	remotePeer.AttachChan(this.NetChan)
	this.ReplacePeer(remotePeer)

	go remotePeer.Link.Rx()
	this.protocol.HandleSystemMessage(this, p2p.PeerConnected{Info: remotePeer.Info})
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

		go func() {
			if err := this.handleClientConnection(conn); err != nil {
				log.Warnf("[p2p] client connect error: %s", err)
				_ = conn.Close()
			}
		}()
	}
}

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() uint {
	return this.connCtrl.OutboundsCount()
}

//check own network address
func (this *NetServer) IsOwnAddress(addr string) bool {
	return addr == this.connCtrl.OwnAddress()
}

func createPeer(info *peer.PeerInfo, conn net.Conn) *peer.Peer {
	remotePeer := peer.NewPeer()
	remotePeer.SetInfo(info)
	remotePeer.Link.UpdateRXTime(time.Now())
	remotePeer.Link.SetAddr(conn.RemoteAddr().String())
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(info.Id)

	return remotePeer
}

func (ns *NetServer) ConnectController() *connect_controller.ConnectController {
	return ns.connCtrl
}

func (ns *NetServer) Protocol() p2p.Protocol {
	return ns.protocol
}

func (this *NetServer) SendTo(p common.PeerId, msg types.Message) {
	peer := this.GetPeer(p)
	if peer != nil {
		this.Send(peer, msg)
	}
}
