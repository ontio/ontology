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
	"github.com/ontio/ontology/common/config"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	conn "github.com/ontio/ontology/p2pserver/link"
	"github.com/ontio/ontology/p2pserver/message/types"
)

// PeerCom provides the basic information of a peer
type PeerCom struct {
	id             uint64
	version        uint32
	services       uint64
	relay          bool
	httpInfoPort   uint16
	syncPort       map[byte]uint16
	consPort       map[byte]uint16
	height         uint64
}

func (this *PeerCom) Init() {
	if config.DefConfig.P2PNode.TransportType == common.LegacyTSPType {
		this.syncPort = make(map[byte]uint16, 1)
		this.consPort = make(map[byte]uint16, 1)
	}else {
		this.syncPort = make(map[byte]uint16, 2)
		this.consPort = make(map[byte]uint16, 2)
	}
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
func (this *PeerCom) SetSyncPort(port uint16, tspType byte) {
	this.syncPort[tspType] = port
}

// GetSyncPort returns a peer's sync port
func (this *PeerCom) GetSyncPort(tspType byte) uint16 {
	return this.syncPort[tspType]
}

// SetConsPort sets a peer's consensus port
func (this *PeerCom) SetConsPort(port uint16, tspType byte) {
	this.consPort[tspType] = port
}

// GetConsPort returns a peer's consensus port
func (this *PeerCom) GetConsPort(tspType byte) uint16 {
	return this.consPort[tspType]
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

//Peer represent the node in p2p
type Peer struct {
	base             PeerCom
	cap              [32]byte
	SyncLink         map[byte]*conn.Link
	ConsLink         map[byte]*conn.Link
	syncState        map[byte]*uint32
	consState        map[byte]*uint32
	txnCnt           map[byte]uint64
	rxTxnCnt         map[byte]uint64
	connLock         sync.RWMutex
	transportType    byte
}

//NewPeer return new peer without publickey initial
func NewPeer() *Peer {
	p := &Peer{}

	p.base.Init()

	if config.DefConfig.P2PNode.TransportType == common.LegacyTSPType {
		p.SyncLink  = map[byte]*conn.Link{common.LegacyTSPType:conn.NewLink()}
		p.ConsLink  = map[byte]*conn.Link{common.LegacyTSPType:conn.NewLink()}

		syncSL := new(uint32)
		consSL := new(uint32)
		*syncSL = common.INIT
		*consSL = common.INIT
		p.syncState = map[byte]*uint32{common.LegacyTSPType:syncSL}
		p.consState = map[byte]*uint32{common.LegacyTSPType:consSL}

		p.txnCnt   = make(map[byte]uint64, 1)
		p.rxTxnCnt = make(map[byte]uint64, 1)
	}else {
		p.SyncLink  = map[byte]*conn.Link{common.LegacyTSPType:conn.NewLink(), config.DefConfig.P2PNode.TransportType:conn.NewLink()}
		p.ConsLink  = map[byte]*conn.Link{common.LegacyTSPType:conn.NewLink(), config.DefConfig.P2PNode.TransportType:conn.NewLink()}

		syncSL := new(uint32)
		syncS  := new(uint32)
		consSL := new(uint32)
		consS  := new(uint32)
		*syncSL = common.INIT
		*syncS  = common.INIT
		*consSL = common.INIT
		*consS  = common.INIT
		p.syncState = map[byte]*uint32{common.LegacyTSPType:syncSL, config.DefConfig.P2PNode.TransportType:syncS}
		p.consState = map[byte]*uint32{common.LegacyTSPType:consSL, config.DefConfig.P2PNode.TransportType:consS}

		p.txnCnt   = make(map[byte]uint64, 2)
		p.rxTxnCnt = make(map[byte]uint64, 2)
	}

	runtime.SetFinalizer(p, rmPeer)
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debugf("[p2p]Remove unused peer: %d", p.GetID())
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node info:")
	log.Debug("[p2p]\t syncState = ", this.syncState)
	log.Debug("[p2p]\t consState = ", this.consState)
	log.Debug("[p2p]\t id = ", this.GetID())
	log.Debug("[p2p]\t addr = ", this.SyncLink[this.transportType].GetAddr())
	log.Debug("[p2p]\t cap = ", this.cap)
	log.Debug("[p2p]\t version = ", this.GetVersion())
	log.Debug("[p2p]\t services = ", this.GetServices())
	log.Debug("[p2p]\t syncPort = ", this.GetSyncPort(this.transportType))
	log.Debug("[p2p]\t consPort = ", this.GetConsPort(this.transportType))
	log.Debug("[p2p]\t relay = ", this.GetRelay())
	log.Debug("[p2p]\t height = ", this.GetHeight())
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
func (this *Peer) GetConsConn(tspType byte) *conn.Link {
	return this.ConsLink[tspType]
}

//SetConsConn set consensue link to peer
func (this *Peer) SetConsConn(consLink *conn.Link, tspType byte) {
	this.ConsLink[tspType] = consLink
}

//GetSyncState return sync state
func (this *Peer) GetSyncState(tspType byte) uint32 {
	return *(this.syncState[tspType])
}

//SetSyncState set sync state to peer
func (this *Peer) SetSyncState(state uint32, tspType byte) {
	atomic.StoreUint32(this.syncState[tspType], state)
}

//GetConsState return peer`s consensus state
func (this *Peer) GetConsState(tspType byte) uint32 {
	return *(this.consState[tspType])
}

//SetConsState set consensus state to peer
func (this *Peer) SetConsState(state uint32, tspType byte) {
	atomic.StoreUint32(this.consState[tspType], state)
}

//GetSyncPort return peer`s sync port
func (this *Peer) GetSyncPort(tspType byte) uint16 {
	return this.SyncLink[tspType].GetPort()
}

//GetConsPort return peer`s consensus port
func (this *Peer) GetConsPort(tspType byte) uint16 {
	return this.ConsLink[tspType].GetPort()
}

//SetConsPort set peer`s consensus port
func (this *Peer) SetConsPort(port uint16, tspType byte) {
	this.ConsLink[tspType].SetPort(port)
}

//GetTransportType return transport type
func (this *Peer) GetTransportType() byte {
	return this.transportType
}

//SetTransportType set transport type to peer
func (this *Peer) SetTransportType(tspType byte) {
	this.transportType = tspType
}

//SendToSync call sync link to send buffer
func (this *Peer) SendToSync(msg types.Message, tspType byte) error {
	if this.SyncLink[tspType] != nil && this.SyncLink[tspType].Valid() {
		return this.SyncLink[tspType].Tx(msg, tspType)
	}
	return errors.New("[p2p]sync link invalid")
}

//SendToCons call consensus link to send buffer
func (this *Peer) SendToCons(msg types.Message, tspType byte) error {
	if this.ConsLink[tspType] != nil && this.ConsLink[tspType].Valid() {
		log.Tracef("TX msg %s by transport %s", msg.CmdType(), common.GetTransportTypeString(tspType))
		return this.ConsLink[tspType].Tx(msg, tspType)
	} else if this.ConsLink[tspType] == nil {
		log.Tracef("peer %d ConsLink nil for transport %s", this.GetID(), common.GetTransportTypeString(tspType))
	}else {
		log.Tracef("peer %d ConsLink(state:%d) invalid for transport %s", this.GetID(), this.GetConsState(tspType), common.GetTransportTypeString(tspType))
	}
	return errors.New("[p2p]cons link invalid")
}

//CloseSync halt sync connection
func (this *Peer) CloseSync(tspType byte) {
	this.SetSyncState(common.INACTIVITY, tspType)

	conn := this.SyncLink[tspType].GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()
	}
	this.connLock.Unlock()
}

//CloseCons halt consensus connection
func (this *Peer) CloseCons(tspType byte) {
	this.SetConsState(common.INACTIVITY, tspType)
	conn := this.ConsLink[tspType].GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()

	}
	this.connLock.Unlock()
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
func (this *Peer) GetTimeStamp(tspType byte) int64 {
	return this.SyncLink[tspType].GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime(tspType byte) time.Time {
	return this.SyncLink[tspType].GetRXTime()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr(tspType byte) string {
	return this.SyncLink[tspType].GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16(tspType byte) ([16]byte, error) {
	var result [16]byte
	addrIp, err := common.ParseIPAddr(this.GetAddr(tspType))
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Warn("[p2p]parse ip address error\n", this.GetAddr(tspType))
		return result, errors.New("[p2p]parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

//AttachSyncChan set msg chan to sync link
func (this *Peer) AttachSyncChan(msgchan chan *types.RecvMessage, tspType byte) {
	this.SyncLink[tspType].SetChan(msgchan)
}

//AttachConsChan set msg chan to consensus link
func (this *Peer) AttachConsChan(msgchan chan *types.RecvMessage, tspType byte) {
	this.ConsLink[tspType].SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(msg types.Message, isConsensus bool, tspType byte) error {
	if isConsensus && this.ConsLink[tspType].Valid() {
		return this.SendToCons(msg, tspType)
	}
	return this.SendToSync(msg, tspType)
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
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64, tspType byte) {

	this.SyncLink[tspType].UpdateRXTime(t)
	this.base.SetID(nonce)
	this.base.SetVersion(version)
	this.base.SetServices(services)
	this.base.SetSyncPort(syncPort, tspType)
	this.base.SetConsPort(consPort, tspType)
	this.SyncLink[tspType].SetPort(syncPort)
	this.ConsLink[tspType].SetPort(consPort)
	if relay == 0 {
		this.base.SetRelay(false)
	} else {
		this.base.SetRelay(true)
	}
	this.SetHeight(uint64(height))
	this.SetTransportType(tspType)
}
