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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/p2pserver/common"
	conn "github.com/ontio/ontology/p2pserver/link"
	//"github.com/ontio/ontology/p2pserver/message"
)

type Peer struct {
	SyncLink                 *conn.Link
	ConsLink                 *conn.Link
	syncState                uint32
	consState                uint32
	id                       uint64
	version                  uint32
	cap                      [32]byte
	services                 uint64
	relay                    bool
	height                   uint64
	txnCnt                   uint64
	rxTxnCnt                 uint64
	httpInfoPort             uint16
	publicKey                keypair.PublicKey
	chF                      chan func() error // Channel used to operate the node without lock
	peerDisconnectSubscriber events.Subscriber
	Np                       *NbrPeers
}

func (p *Peer) backend() {
	for f := range p.chF {
		f()
	}
}

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

func (p *Peer) InitPeer(pubKey keypair.PublicKey) error {
	p.version = common.PROTOCOL_VERSION
	if config.Parameters.NodeType == common.SERVICE_NODE_NAME {
		p.services = uint64(common.SERVICE_NODE)
	} else if config.Parameters.NodeType == common.VERIFY_NODE_NAME {
		p.services = uint64(common.VERIFY_NODE)
	}

	if config.Parameters.NodeConsensusPort == 0 || config.Parameters.NodePort == 0 ||
		config.Parameters.NodeConsensusPort == config.Parameters.NodePort {
		log.Error("Network port invalid, please check config.json")
		return errors.New("Invalid port")
	}
	p.SyncLink.SetPort(config.Parameters.NodePort)
	p.ConsLink.SetPort(config.Parameters.NodeConsensusPort)

	p.SyncLink.SetPort(config.Parameters.NodePort)
	p.ConsLink.SetPort(config.Parameters.NodeConsensusPort)

	p.relay = true

	key := keypair.SerializePublicKey(pubKey)

	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(p.id))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(fmt.Sprintf("Init peer ID to 0x%x", p.id))
	p.Np = &NbrPeers{}
	p.Np.init()

	p.publicKey = pubKey
	p.SyncLink.SetID(p.id)
	p.ConsLink.SetID(p.id)
	return nil
}

func rmPeer(p *Peer) {
	log.Debug(fmt.Sprintf("Remove unused peer: 0x%0x", p.id))
}
func (p *Peer) DumpInfo() {
	log.Info("Node info:")
	log.Info("\t syncState = ", p.syncState)
	log.Info("\t consState = ", p.consState)
	log.Info(fmt.Sprintf("\t id = 0x%x", p.id))
	log.Info("\t addr = ", p.SyncLink.GetAddr())
	log.Info("\t cap = ", p.cap)
	log.Info("\t version = ", p.version)
	log.Info("\t services = ", p.services)
	log.Info("\t syncPort = ", p.SyncLink.GetPort())
	log.Info("\t consPort = ", p.ConsLink.GetPort())
	log.Info("\t relay = ", p.relay)
	log.Info("\t height = ", p.height)
}
func (p *Peer) SetBookKeeperAddr(pubKey keypair.PublicKey) {
	p.publicKey = pubKey
}
func (p *Peer) GetPubKey() keypair.PublicKey {
	return p.publicKey
}
func (p *Peer) GetVersion() uint32 {
	return p.version
}
func (p *Peer) GetHeight() uint64 {
	return p.height
}
func (p *Peer) SetHeight(height uint64) {
	p.height = height
}
func (p *Peer) GetConsConn() *conn.Link {
	return p.ConsLink
}
func (p *Peer) SetConsConn(consLink *conn.Link) {
	p.ConsLink = consLink
}
func (p *Peer) GetSyncState() uint32 {
	return p.syncState
}
func (p *Peer) SetSyncState(state uint32) {
	atomic.StoreUint32(&(p.syncState), state)
}
func (p *Peer) GetConsState() uint32 {
	return p.consState
}
func (p *Peer) SetConsState(state uint32) {
	atomic.StoreUint32(&(p.consState), state)
}
func (p *Peer) GetSyncPort() uint16 {
	return p.SyncLink.GetPort()
}
func (p *Peer) GetConsPort() uint16 {
	return p.ConsLink.GetPort()
}
func (p *Peer) SetConsPort(port uint16) {
	p.ConsLink.SetPort(port)
}
func (p *Peer) SendToSync(buf []byte) {
	p.SyncLink.Tx(buf)
}
func (p *Peer) SendToCons(buf []byte) {
	p.ConsLink.Tx(buf)
}
func (p *Peer) CloseSync() {
	p.SetSyncState(common.INACTIVITY)
	conn := p.SyncLink.GetConn()
	if conn != nil {
		conn.Close()
	}

}
func (p *Peer) CloseCons() {
	p.SetConsState(common.INACTIVITY)
	conn := p.ConsLink.GetConn()
	if conn != nil {
		conn.Close()
	}
}

func (p *Peer) GetID() uint64 {
	return p.id
}
func (p *Peer) GetRelay() bool {
	return p.relay
}
func (p *Peer) GetServices() uint64 {
	return p.services
}
func (p *Peer) GetTimeStamp() int64 {
	return p.SyncLink.GetRXTime().UnixNano()
}
func (p *Peer) GetContactTime() time.Time {
	return p.SyncLink.GetRXTime()
}
func (p *Peer) GetAddr() string {
	return p.SyncLink.GetAddr()
}
func (p *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	ip := net.ParseIP(p.GetAddr()).To16()
	if ip == nil {
		log.Error("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (p *Peer) AttachSyncChan(msgchan chan common.MsgPayload) {
	p.SyncLink.SetChan(msgchan)
}
func (p *Peer) AttachConsChan(msgchan chan common.MsgPayload) {
	p.ConsLink.SetChan(msgchan)
}

func (p *Peer) DelNbrNode(id uint64) (*Peer, bool) {
	return p.Np.DelNbrNode(id)
}

func (p *Peer) Send(buf []byte, isConsensus bool) error {
	if isConsensus && p.ConsLink.Valid() {
		return p.ConsLink.Tx(buf)
	}
	return p.SyncLink.Tx(buf)
}

func (p *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		p.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		p.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

func (p *Peer) GetHttpInfoState() bool {
	return p.cap[common.HTTP_INFO_FLAG] == 1
}

func (p *Peer) GetHttpInfoPort() uint16 {
	return p.httpInfoPort
}

func (p *Peer) SetHttpInfoPort(port uint16) {
	p.httpInfoPort = port
}

//SetBookkeeperAddr set peer`s publickey
func (p *Peer) SetBookkeeperAddr(pk *keypair.PublicKey) {
	p.publicKey = pk
}

//UpdateInfo update peer`s information
func (p *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64) {

	p.SyncLink.UpdateRXTime(t)
	p.id = nonce
	p.version = version
	p.services = services
	p.SyncLink.SetPort(syncPort)
	p.ConsLink.SetPort(consPort)
	if relay == 0 {
		p.relay = false
	} else {
		p.relay = true
	}
	p.height = uint64(height)
}

//AddNbrNode add peer to nbr peer list
func (p *Peer) AddNbrNode(remotePeer *Peer) {
	p.Np.AddNbrNode(remotePeer)
}
