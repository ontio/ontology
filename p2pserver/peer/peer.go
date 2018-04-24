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
func (pc *PeerCom) SetID(id uint64) {
	pc.id = id
}

// GetID returns a peer's id
func (pc *PeerCom) GetID() uint64 {
	return pc.id
}

// SetVersion sets a peer's version
func (pc *PeerCom) SetVersion(version uint32) {
	pc.version = version
}

// GetVersion returns a peer's version
func (pc *PeerCom) GetVersion() uint32 {
	return pc.version
}

// SetServices sets a peer's services
func (pc *PeerCom) SetServices(services uint64) {
	pc.services = services
}

// GetServices returns a peer's services
func (pc *PeerCom) GetServices() uint64 {
	return pc.services
}

// SerRelay sets a peer's relay
func (pc *PeerCom) SetRelay(relay bool) {
	pc.relay = relay
}

// GetRelay returns a peer's relay
func (pc *PeerCom) GetRelay() bool {
	return pc.relay
}

// SetSyncPort sets a peer's sync port
func (pc *PeerCom) SetSyncPort(port uint16) {
	pc.syncPort = port
}

// GetSyncPort returns a peer's sync port
func (pc *PeerCom) GetSyncPort() uint16 {
	return pc.syncPort
}

// SetConsPort sets a peer's consensus port
func (pc *PeerCom) SetConsPort(port uint16) {
	pc.consPort = port
}

// GetConsPort returns a peer's consensus port
func (pc *PeerCom) GetConsPort() uint16 {
	return pc.consPort
}

// SetHttpInfoPort sets a peer's http info port
func (pc *PeerCom) SetHttpInfoPort(port uint16) {
	pc.httpInfoPort = port
}

// GetHttpInfoPort returns a peer's http info port
func (pc *PeerCom) GetHttpInfoPort() uint16 {
	return pc.httpInfoPort
}

// SetHeight sets a peer's height
func (pc *PeerCom) SetHeight(height uint64) {
	pc.height = height
}

// GetHeight returns a peer's height
func (pc *PeerCom) GetHeight() uint64 {
	return pc.height
}

// SetPubKey sets a peer's public key
func (pc *PeerCom) SetPubKey(pubKey keypair.PublicKey) {
	pc.publicKey = pubKey
}

// GetPubKey returns a peer's public key
func (pc *PeerCom) GetPubKey() keypair.PublicKey {
	return pc.publicKey
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
func (p *Peer) backend() {
	for f := range p.chF {
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
func (p *Peer) DumpInfo() {
	log.Info("Node info:")
	log.Info("\t syncState = ", p.syncState)
	log.Info("\t consState = ", p.consState)
	log.Info(fmt.Sprintf("\t id = 0x%x", p.GetID()))
	log.Info("\t addr = ", p.SyncLink.GetAddr())
	log.Info("\t cap = ", p.cap)
	log.Info("\t version = ", p.GetVersion())
	log.Info("\t services = ", p.GetServices())
	log.Info("\t syncPort = ", p.GetSyncPort())
	log.Info("\t consPort = ", p.GetConsPort())
	log.Info("\t relay = ", p.GetRelay())
	log.Info("\t height = ", p.GetHeight())
}

//SetBookKeeperAddr set pubKey to peer
func (p *Peer) SetBookKeeperAddr(pubKey keypair.PublicKey) {
	p.base.SetPubKey(pubKey)
}

//GetPubKey return publickey of peer
func (p *Peer) GetPubKey() keypair.PublicKey {
	return p.base.GetPubKey()
}

//GetVersion return peer`s version
func (p *Peer) GetVersion() uint32 {
	return p.base.GetVersion()
}

//GetHeight return peer`s block height
func (p *Peer) GetHeight() uint64 {
	return p.base.GetHeight()
}

//SetHeight set height to peer
func (p *Peer) SetHeight(height uint64) {
	p.base.SetHeight(height)
}

//GetConsConn return consensus link
func (p *Peer) GetConsConn() *conn.Link {
	return p.ConsLink
}

//SetConsConn set consensue link to peer
func (p *Peer) SetConsConn(consLink *conn.Link) {
	p.ConsLink = consLink
}

//GetSyncState return sync state
func (p *Peer) GetSyncState() uint32 {
	return p.syncState
}

//SetSyncState set sync state to peer
func (p *Peer) SetSyncState(state uint32) {
	atomic.StoreUint32(&(p.syncState), state)
}

//GetConsState return peer`s consensus state
func (p *Peer) GetConsState() uint32 {
	return p.consState
}

//SetConsState set consensus state to peer
func (p *Peer) SetConsState(state uint32) {
	atomic.StoreUint32(&(p.consState), state)
}

//GetSyncPort return peer`s sync port
func (p *Peer) GetSyncPort() uint16 {
	return p.SyncLink.GetPort()
}

//GetConsPort return peer`s consensus port
func (p *Peer) GetConsPort() uint16 {
	return p.ConsLink.GetPort()
}

//SetConsPort set peer`s consensus port
func (p *Peer) SetConsPort(port uint16) {
	p.ConsLink.SetPort(port)
}

//SendToSync call sync link to send buffer
func (p *Peer) SendToSync(buf []byte) {
	if p.SyncLink != nil && p.SyncLink.Valid() {
		p.SyncLink.Tx(buf)
	}

}

//SendToCons call consensus link to send buffer
func (p *Peer) SendToCons(buf []byte) {
	if p.ConsLink != nil && p.ConsLink.Valid() {
		p.ConsLink.Tx(buf)
	}
}

//CloseSync halt sync connection
func (p *Peer) CloseSync() {
	p.SetSyncState(common.INACTIVITY)
	conn := p.SyncLink.GetConn()
	if conn != nil {
		conn.Close()
	}

}

//CloseCons halt consensus connection
func (p *Peer) CloseCons() {
	p.SetConsState(common.INACTIVITY)
	conn := p.ConsLink.GetConn()
	if conn != nil {
		conn.Close()
	}
}

//GetID return peer`s id
func (p *Peer) GetID() uint64 {
	return p.base.GetID()
}

//GetRelay return peer`s relay state
func (p *Peer) GetRelay() bool {
	return p.base.GetRelay()
}

//GetServices return peer`s service state
func (p *Peer) GetServices() uint64 {
	return p.base.GetServices()
}

//GetTimeStamp return peer`s latest contact time in ticks
func (p *Peer) GetTimeStamp() int64 {
	return p.SyncLink.GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (p *Peer) GetContactTime() time.Time {
	return p.SyncLink.GetRXTime()
}

//GetAddr return peer`s sync link address
func (p *Peer) GetAddr() string {
	return p.SyncLink.GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (p *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := parseIPAddr(p.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", p.GetAddr())
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

//AttachSyncChan set msg chan to sync link
func (p *Peer) AttachSyncChan(msgchan chan *common.MsgPayload) {
	p.SyncLink.SetChan(msgchan)
}

//AttachConsChan set msg chan to consensus link
func (p *Peer) AttachConsChan(msgchan chan *common.MsgPayload) {
	p.ConsLink.SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (p *Peer) Send(buf []byte, isConsensus bool) error {
	if isConsensus && p.ConsLink.Valid() {
		return p.ConsLink.Tx(buf)
	}
	return p.SyncLink.Tx(buf)
}

//SetHttpInfoState set peer`s httpinfo state
func (p *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		p.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		p.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

//GetHttpInfoState return peer`s httpinfo state
func (p *Peer) GetHttpInfoState() bool {
	return p.cap[common.HTTP_INFO_FLAG] == 1
}

//GetHttpInfoPort return peer`s httpinfo port
func (p *Peer) GetHttpInfoPort() uint16 {
	return p.base.GetHttpInfoPort()
}

//SetHttpInfoPort set peer`s httpinfo port
func (p *Peer) SetHttpInfoPort(port uint16) {
	p.base.SetHttpInfoPort(port)
}

//UpdateInfo update peer`s information
func (p *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64) {

	p.SyncLink.UpdateRXTime(t)
	p.base.SetID(nonce)
	p.base.SetVersion(version)
	p.base.SetServices(services)
	p.base.SetSyncPort(syncPort)
	p.base.SetConsPort(consPort)
	p.SyncLink.SetPort(syncPort)
	p.ConsLink.SetPort(consPort)
	if relay == 0 {
		p.base.SetRelay(false)
	} else {
		p.base.SetRelay(true)
	}
	p.SetHeight(uint64(height))
}

//parseIPAddr return ip address
func parseIPAddr(s string) (string, error) {
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("Split IP address error")
		return s, errors.New("Split IP address error")
	}
	return s[:i], nil
}
