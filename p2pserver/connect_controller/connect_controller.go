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
package connect_controller

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/handshake"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/scylladb/go-set/strset"
)

const INBOUND_INDEX = 0
const OUTBOUND_INDEX = 1

type connectedPeer struct {
	connectId uint64
	addr      string
	peer      *peer.PeerInfo
}

type ConnectController struct {
	ConnCtrlOption

	selfId   *common.PeerKeyId
	peerInfo *peer.PeerInfo

	mutex       sync.Mutex
	inoutbounds [2]*strset.Set // in/outbounds address list
	connecting  *strset.Set
	peers       map[common.PeerId]*connectedPeer // all connected peers

	ownAddr       string
	nextConnectId uint64
}

func NewConnectController(peerInfo *peer.PeerInfo, keyid *common.PeerKeyId,
	option ConnCtrlOption) *ConnectController {
	control := &ConnectController{
		ConnCtrlOption: option,
		selfId:         keyid,
		peerInfo:       peerInfo,
		inoutbounds:    [2]*strset.Set{strset.New(), strset.New()},
		connecting:     strset.New(),
		peers:          make(map[common.PeerId]*connectedPeer),
	}
	// put domain to the end
	sort.Slice(control.ReservedPeers, func(i, j int) bool {
		return net.ParseIP(control.ReservedPeers[i]) != nil
	})

	return control
}

func (self *ConnectController) OwnAddress() string {
	return self.ownAddr
}

func (self *ConnectController) getConnectId() uint64 {
	return atomic.AddUint64(&self.nextConnectId, 1)
}

func (self *ConnectController) hasInbound(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.inoutbounds[INBOUND_INDEX].Has(addr)
}

func (self *ConnectController) OutboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[OUTBOUND_INDEX].Size())
}

func (self *ConnectController) InboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[INBOUND_INDEX].Size())
}

func (self *ConnectController) isBoundFull(index int) bool {
	count := self.boundsCount(index)
	if index == INBOUND_INDEX {
		return count >= self.MaxConnInBound
	}
	return count >= self.MaxConnOutBound
}

func (self *ConnectController) boundsCount(index int) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[index].Size())
}

func (self *ConnectController) hasBoundAddr(addr string, index int) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.inoutbounds[index].Has(addr)
}

func (self *ConnectController) tryAddConnecting(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.connecting.Has(addr) {
		return false
	}
	self.connecting.Add(addr)

	return true
}

func (self *ConnectController) removeConnecting(addr string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.connecting.Remove(addr)
}

func (self *ConnectController) reserveEnabled() bool {
	return len(self.ReservedPeers) > 0
}

// remoteAddr format 192.168.1.1:61234
func (self *ConnectController) inReserveList(remoteIPPort string) bool {
	// 192.168.1.1 in reserve list, 192.168.1.111:61234 and 192.168.1.11:61234 can connect in if we are using prefix matching
	// so get this IP to do fully match
	remoteAddr, _, err := net.SplitHostPort(remoteIPPort)
	if err != nil {
		return false
	}
	// we don't load domain in start because we consider domain's A/AAAA record may change sometimes
	for _, curIPOrName := range self.ReservedPeers {
		curIPs, err := net.LookupHost(curIPOrName)
		if err != nil {
			continue
		}
		for _, digIP := range curIPs {
			if digIP == remoteAddr {
				return true
			}
		}
	}

	return false
}

func (self *ConnectController) checkReservedPeers(remoteAddr string) error {
	if !self.reserveEnabled() || self.inReserveList(remoteAddr) {
		return nil
	}
	return fmt.Errorf("the remote addr: %s not in reserved list", remoteAddr)
}

func (self *ConnectController) getInboundCountWithIp(ip string) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	var count uint
	self.inoutbounds[INBOUND_INDEX].Each(func(addr string) bool {
		ipRecord, _ := common.ParseIPAddr(addr)
		if ipRecord == ip {
			count += 1
		}

		return true
	})

	return count
}

func (self *ConnectController) AcceptConnect(conn net.Conn) (*peer.PeerInfo, net.Conn, error) {
	addr := conn.RemoteAddr().String()
	err := self.beforeHandshakeCheck(addr, INBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeServer(self.peerInfo, self.selfId, conn)
	if err != nil {
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, addr)
	if err != nil {
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, INBOUND_INDEX)

	log.Infof("inbound peer %s connected, %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

//Connect used to connect net address under sync or cons mode
// need call Peer.Close to clean up resource.
func (self *ConnectController) Connect(addr string) (*peer.PeerInfo, net.Conn, error) {
	err := self.beforeHandshakeCheck(addr, OUTBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	if !self.tryAddConnecting(addr) {
		return nil, nil, fmt.Errorf("node exist in connecting list: %s", addr)
	}
	defer self.removeConnecting(addr)

	conn, err := self.dialer.Dial(addr)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeClient(self.peerInfo, self.selfId, conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, OUTBOUND_INDEX)

	log.Infof("outbound peer %s connected. %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

func (self *ConnectController) afterHandshakeCheck(remotePeer *peer.PeerInfo, remoteAddr string) error {
	if err := self.isHandWithSelf(remotePeer, remoteAddr); err != nil {
		return err
	}

	return self.checkPeerIdAndIP(remotePeer, remoteAddr)
}

func (self *ConnectController) beforeHandshakeCheck(addr string, index int) error {
	err := self.checkReservedPeers(addr)
	if err != nil {
		return err
	}

	if self.hasBoundAddr(addr, index) {
		return fmt.Errorf("peer %s already in connection records", addr)
	}

	if self.ownAddr == addr {
		return fmt.Errorf("connecting with self address %s", addr)
	}

	if self.isBoundFull(index) {
		return fmt.Errorf("[p2p] bound %d connections reach max limit", index)
	}
	if index == INBOUND_INDEX {
		remoteIp, err := common.ParseIPAddr(addr)
		if err != nil {
			return fmt.Errorf("[p2p]parse ip error %v", err.Error())
		}
		connNum := self.getInboundCountWithIp(remoteIp)
		if connNum >= self.MaxConnInBoundPerIP {
			return fmt.Errorf("connections(%d) with ip(%s) has reach max limit(%d), "+
				"conn closed", connNum, remoteIp, self.MaxConnInBoundPerIP)
		}
	}

	return nil
}

func (self *ConnectController) isHandWithSelf(remotePeer *peer.PeerInfo, remoteAddr string) error {
	addrIp, err := common.ParseIPAddr(remoteAddr)
	if err != nil {
		log.Warn(err)
		return err
	}
	nodeAddr := addrIp + ":" + strconv.Itoa(int(remotePeer.Port))
	if remotePeer.Id.ToUint64() == self.selfId.Id.ToUint64() {
		self.ownAddr = nodeAddr
		return fmt.Errorf("the node handshake with itself: %s", remoteAddr)
	}

	return nil
}

func (self *ConnectController) getPeer(kid common.PeerId) *connectedPeer {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	p := self.peers[kid]
	return p
}

func (self *ConnectController) savePeer(conn net.Conn, p *peer.PeerInfo, index int) net.Conn {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	addr := conn.RemoteAddr().String()
	self.inoutbounds[index].Add(addr)

	cid := self.getConnectId()
	self.peers[p.Id] = &connectedPeer{
		connectId: cid,
		addr:      addr,
		peer:      p,
	}

	return &Conn{
		Conn:       conn,
		connectId:  cid,
		kid:        p.Id,
		addr:       addr,
		boundIndex: index,
		controller: self,
	}
}

func (self *ConnectController) removePeer(conn *Conn) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.inoutbounds[conn.boundIndex].Remove(conn.addr)

	p := self.peers[conn.kid]
	if p == nil || p.peer == nil {
		log.Fatalf("connection %s not in controller", conn.kid.ToHexString())
	} else if p.connectId == conn.connectId { // connection not replaced
		delete(self.peers, conn.kid)
	}
}

// if connection with peer.Kid exist, but has different IP, return error
func (self *ConnectController) checkPeerIdAndIP(peer *peer.PeerInfo, addr string) error {
	oldPeer := self.getPeer(peer.Id)
	if oldPeer == nil {
		return nil
	}

	ipOld, err := common.ParseIPAddr(oldPeer.addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]exist peer ip format is wrong %s", oldPeer.addr)
		log.Fatal(err)
		return err
	}
	ipNew, err := common.ParseIPAddr(addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]connecting peer ip format is wrong %s, close", addr)
		log.Fatal(err)
		return err
	}

	if ipNew != ipOld {
		err := fmt.Errorf("[createPeer]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
		log.Warn(err)
		return err
	}

	return nil
}
