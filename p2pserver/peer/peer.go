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
	"sync"
	"sync/atomic"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	conn "github.com/ontio/ontology/p2pserver/link"
	"github.com/ontio/ontology/p2pserver/message/types"
)

// PeerInfo provides the basic information of a peer
type PeerInfo struct {
	Id           kbucket.KadId
	Version      uint32
	Services     uint64
	Relay        bool
	HttpInfoPort uint16
	Port         uint16
	Height       uint64
	SoftVersion  string
}

func NewPeerInfo(id kbucket.KadId, version uint32, services uint64, relay bool, httpInfoPort uint16,
	port uint16, height uint64, softVersion string) *PeerInfo {
	return &PeerInfo{
		Id:           id,
		Version:      version,
		Services:     services,
		Relay:        relay,
		HttpInfoPort: httpInfoPort,
		Port:         port,
		Height:       height,
		SoftVersion:  softVersion,
	}
}

//Peer represent the node in p2p
type Peer struct {
	info      *PeerInfo
	cap       [32]byte
	Link      *conn.Link
	linkState uint32
	txnCnt    uint64
	rxTxnCnt  uint64
	connLock  sync.RWMutex
}

//NewPeer return new peer without publickey initial
func NewPeer() *Peer {
	p := &Peer{
		linkState: common.INIT,
		info:      &PeerInfo{},
		Link:      conn.NewLink(),
	}
	runtime.SetFinalizer(p, rmPeer)
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debugf("[p2p]Remove unused peer: %d", p.GetID())
}

func (self *Peer) SetInfo(info *PeerInfo) {
	self.info = info
}

func (self *PeerInfo) String() string {
	return fmt.Sprintf("id=%s, version=%s", self.Id.ToHexString(), self.SoftVersion)
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node info:")
	log.Debug("[p2p]\t linkState = ", this.linkState)
	log.Debug("[p2p]\t id = ", this.GetID())
	log.Debug("[p2p]\t addr = ", this.Link.GetAddr())
	log.Debug("[p2p]\t cap = ", this.cap)
	log.Debug("[p2p]\t version = ", this.GetVersion())
	log.Debug("[p2p]\t services = ", this.GetServices())
	log.Debug("[p2p]\t port = ", this.GetPort())
	log.Debug("[p2p]\t relay = ", this.GetRelay())
	log.Debug("[p2p]\t height = ", this.GetHeight())
	log.Debug("[p2p]\t softVersion = ", this.GetSoftVersion())
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.info.Version
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.info.Height
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.info.Height = height
}

//GetState return sync state
func (this *Peer) GetState() uint32 {
	return this.linkState
}

//SetState set sync state to peer
func (this *Peer) SetState(state uint32) {
	atomic.StoreUint32(&(this.linkState), state)
}

//GetPort return peer`s sync port
func (this *Peer) GetPort() uint16 {
	return this.Link.GetPort()
}

//SendTo call sync link to send buffer
func (this *Peer) SendRaw(msgType string, msgPayload []byte) error {
	if this.Link != nil && this.Link.Valid() {
		return this.Link.SendRaw(msgPayload)
	}
	return errors.New("[p2p]sync link invalid")
}

//Close halt sync connection
func (this *Peer) Close() {
	this.SetState(common.INACTIVITY)
	conn := this.Link.GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()
	}
	this.connLock.Unlock()
}

//GetID return peer`s id
func (this *Peer) GetID() uint64 {
	return this.info.Id.ToUint64()
}

func (this *Peer) GetKId() kbucket.KadId {
	return this.info.Id
}

func (this *Peer) SetKId(id kbucket.KadId) {
	this.info.Id = id
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.info.Relay
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.info.Services
}

//GetTimeStamp return peer`s latest contact time in ticks
func (this *Peer) GetTimeStamp() int64 {
	return this.Link.GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime() time.Time {
	return this.Link.GetRXTime()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() string {
	return this.Link.GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := common.ParseIPAddr(this.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Warn("[p2p]parse ip address error\n", this.GetAddr())
		return result, errors.New("[p2p]parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (this *Peer) GetSoftVersion() string {
	return this.info.SoftVersion
}

//AttachChan set msg chan to sync link
func (this *Peer) AttachChan(msgchan chan *types.MsgPayload) {
	this.Link.SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	return this.SendRaw(msg.CmdType(), sink.Bytes())
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
	return this.info.HttpInfoPort
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetHttpInfoPort(port uint16) {
	this.info.HttpInfoPort = port
}

//UpdateInfo update peer`s information
func (this *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, kid kbucket.KadId, relay uint8, height uint64, softVer string) {
	this.info.Id = kid
	this.info.Version = version
	this.info.Services = services
	this.info.Port = syncPort
	this.info.SoftVersion = softVer
	this.info.Relay = relay != 0
	this.info.Height = height

	this.Link.UpdateRXTime(t)
	this.Link.SetPort(syncPort)
}

//func NewPeer(t time.Time, version uint32, services uint64,
//	syncPort uint16, nonce uint64, relay uint8, height uint64, softVer string) *Peer {
//		id := kbucket.PseudoKadIdFromUint64(nonce)
//		peerCom := NewPeerCom(id, version,services, relay,true,syncPort,height,softVer)
//		return &Peer{
//
//		}
//}
