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

package peer

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	actor "github.com/ontio/ontology/p2pserver/actor/req"
	"github.com/ontio/ontology/p2pserver/common"
	conn "github.com/ontio/ontology/p2pserver/link"
)

// PeerCom provides the basic information of a peer
type PeerCom struct {
	id           uint64
	version      uint32
	services     uint64
	relay        bool
	httpInfoPort uint16
	syncPort     uint16
	consPort     uint16
	height       uint64
	publicKey    keypair.PublicKey
}

// SetID sets a peer's id
func (this *PeerCom) SetID(id uint64) {
	this.id = id
}

// GetID returns a peer's id
func (this *PeerCom) GetID() uint64 {
	return this.id
}

// SetVersion sets a peer's version
func (this *PeerCom) SetVersion(version uint32) {
	this.version = version
}

// GetVersion returns a peer's version
func (this *PeerCom) GetVersion() uint32 {
	return this.version
}

// SetServices sets a peer's services
func (this *PeerCom) SetServices(services uint64) {
	this.services = services
}

// GetServices returns a peer's services
func (this *PeerCom) GetServices() uint64 {
	return this.services
}

// SerRelay sets a peer's relay
func (this *PeerCom) SetRelay(relay bool) {
	this.relay = relay
}

// GetRelay returns a peer's relay
func (this *PeerCom) GetRelay() bool {
	return this.relay
}

// SetSyncPort sets a peer's sync port
func (this *PeerCom) SetSyncPort(port uint16) {
	this.syncPort = port
}

// GetSyncPort returns a peer's sync port
func (this *PeerCom) GetSyncPort() uint16 {
	return this.syncPort
}

// SetConsPort sets a peer's consensus port
func (this *PeerCom) SetConsPort(port uint16) {
	this.consPort = port
}

// GetConsPort returns a peer's consensus port
func (this *PeerCom) GetConsPort() uint16 {
	return this.consPort
}

// SetHttpInfoPort sets a peer's http info port
func (this *PeerCom) SetHttpInfoPort(port uint16) {
	this.httpInfoPort = port
}

// GetHttpInfoPort returns a peer's http info port
func (this *PeerCom) GetHttpInfoPort() uint16 {
	return this.httpInfoPort
}

// SetHeight sets a peer's height
func (this *PeerCom) SetHeight(height uint64) {
	this.height = height
}

// GetHeight returns a peer's height
func (this *PeerCom) GetHeight() uint64 {
	return this.height
}

// SetPubKey sets a peer's public key
func (this *PeerCom) SetPubKey(pubKey keypair.PublicKey) {
	this.publicKey = pubKey
}

// GetPubKey returns a peer's public key
func (this *PeerCom) GetPubKey() keypair.PublicKey {
	return this.publicKey
}

//Peer represent the node in p2p
type Peer struct {
	base      PeerCom
	cap       [32]byte
	SyncLink  *conn.Link
	ConsLink  *conn.Link
	syncState uint32
	consState uint32
	txnCnt    uint64
	rxTxnCnt  uint64
	chF       chan func() error
}

//backend run function in backend
func (this *Peer) backend() {
	for f := range this.chF {
		f()
	}
}

//NewPeer return new peer without publickey initial
func NewPeer() *Peer {
	p := &Peer{
		syncState: common.INIT,
		consState: common.INIT,
		chF:       make(chan func() error),
	}
	p.SyncLink = conn.NewLink()
	p.ConsLink = conn.NewLink()

	runtime.SetFinalizer(p, rmPeer)
	go p.backend()
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", p.GetID()))
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Info("Node info:")
	log.Info("\t syncState = ", this.syncState)
	log.Info("\t consState = ", this.consState)
	log.Info("\t id = 0x%x", this.GetID())
	log.Info("\t addr = ", this.SyncLink.GetAddr())
	log.Info("\t cap = ", this.cap)
	log.Info("\t version = ", this.GetVersion())
	log.Info("\t services = ", this.GetServices())
	log.Info("\t syncPort = ", this.GetSyncPort())
	log.Info("\t consPort = ", this.GetConsPort())
	log.Info("\t relay = ", this.GetRelay())
	log.Info("\t height = ", this.GetHeight())
}

//SetBookKeeperAddr set pubKey to peer
func (this *Peer) SetBookKeeperAddr(pubKey keypair.PublicKey) {
	this.base.SetPubKey(pubKey)
}

//GetPubKey return publickey of peer
func (this *Peer) GetPubKey() keypair.PublicKey {
	return this.base.GetPubKey()
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.base.GetHeight()
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

//GetConsConn return consensus link
func (this *Peer) GetConsConn() *conn.Link {
	return this.ConsLink
}

//SetConsConn set consensue link to peer
func (this *Peer) SetConsConn(consLink *conn.Link) {
	this.ConsLink = consLink
}

//GetSyncState return sync state
func (this *Peer) GetSyncState() uint32 {
	return this.syncState
}

//SetSyncState set sync state to peer
func (this *Peer) SetSyncState(state uint32) {
	atomic.StoreUint32(&(this.syncState), state)
	if state == common.ESTABLISH {
		actor.NotifyPeerState(this.GetPubKey(), true)
	}
}

//GetConsState return peer`s consensus state
func (this *Peer) GetConsState() uint32 {
	return this.consState
}

//SetConsState set consensus state to peer
func (this *Peer) SetConsState(state uint32) {
	atomic.StoreUint32(&(this.consState), state)
}

//GetSyncPort return peer`s sync port
func (this *Peer) GetSyncPort() uint16 {
	return this.SyncLink.GetPort()
}

//GetConsPort return peer`s consensus port
func (this *Peer) GetConsPort() uint16 {
	return this.ConsLink.GetPort()
}

//SetConsPort set peer`s consensus port
func (this *Peer) SetConsPort(port uint16) {
	this.ConsLink.SetPort(port)
}

//SendToSync call sync link to send buffer
func (this *Peer) SendToSync(buf []byte) {
	if this.SyncLink != nil && this.SyncLink.Valid() {
		this.SyncLink.Tx(buf)
	}

}

//SendToCons call consensus link to send buffer
func (this *Peer) SendToCons(buf []byte) {
	if this.ConsLink != nil && this.ConsLink.Valid() {
		this.ConsLink.Tx(buf)
	}
}

//CloseSync halt sync connection
func (this *Peer) CloseSync() {
	this.SetSyncState(common.INACTIVITY)
	actor.NotifyPeerState(this.GetPubKey(), false)
	conn := this.SyncLink.GetConn()
	if conn != nil {
		conn.Close()
	}

}

//CloseCons halt consensus connection
func (this *Peer) CloseCons() {
	this.SetConsState(common.INACTIVITY)
	conn := this.ConsLink.GetConn()
	if conn != nil {
		conn.Close()
	}
}

//GetID return peer`s id
func (this *Peer) GetID() uint64 {
	return this.base.GetID()
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.base.GetRelay()
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetTimeStamp return peer`s latest contact time in ticks
func (this *Peer) GetTimeStamp() int64 {
	return this.SyncLink.GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime() time.Time {
	return this.SyncLink.GetRXTime()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() string {
	return this.SyncLink.GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := parseIPAddr(this.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Error("parse ip address error\n", this.GetAddr())
		return result, errors.New("parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

//AttachSyncChan set msg chan to sync link
func (this *Peer) AttachSyncChan(msgchan chan *common.MsgPayload) {
	this.SyncLink.SetChan(msgchan)
}

//AttachConsChan set msg chan to consensus link
func (this *Peer) AttachConsChan(msgchan chan *common.MsgPayload) {
	this.ConsLink.SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(buf []byte, isConsensus bool) error {
	if isConsensus && this.ConsLink.Valid() {
		return this.ConsLink.Tx(buf)
	}
	return this.SyncLink.Tx(buf)
}

//SetHttpInfoState set peer`s httpinfo state
func (this *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		this.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		this.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

//GetHttpInfoState return peer`s httpinfo state
func (this *Peer) GetHttpInfoState() bool {
	return this.cap[common.HTTP_INFO_FLAG] == 1
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetHttpInfoPort(port uint16) {
	this.base.SetHttpInfoPort(port)
}

//UpdateInfo update peer`s information
func (this *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64) {

	this.SyncLink.UpdateRXTime(t)
	this.base.SetID(nonce)
	this.base.SetVersion(version)
	this.base.SetServices(services)
	this.base.SetSyncPort(syncPort)
	this.base.SetConsPort(consPort)
	this.SyncLink.SetPort(syncPort)
	this.ConsLink.SetPort(consPort)
	if relay == 0 {
		this.base.SetRelay(false)
	} else {
		this.base.SetRelay(true)
	}
	this.SetHeight(uint64(height))
}

//parseIPAddr return ip address
func parseIPAddr(s string) (string, error) {
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("split ip address error")
		return s, errors.New("split ip address error")
	}
	return s[:i], nil
}
