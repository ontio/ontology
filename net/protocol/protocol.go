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

package protocol

import (
	"bytes"
	"encoding/binary"
	"time"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology-crypto/keypair"
)

type NodeAddr struct {
	Time     int64
	Services uint64
	IpAddr   [16]byte
	Port     uint16
	ID       uint64 // Unique ID
}

// The node capability type
const (
	VERIFY_NODE  = 1
	SERVICE_NODE = 2
)

const (
	VERIFY_NODE_NAME  = "verify"
	SERVICE_NODE_NAME = "service"
)

const (
	MSG_CMD_LEN           = 12
	CMD_OFFSET            = 4
	CHECKSUM_LEN          = 4
	HASH_LEN              = 32 // hash length in byte
	MSG_HDR_LEN           = 24
	NET_MAGIC             = 0x74746e41
	MAX_BLK_HDR_CNT       = 500
	MAX_INV_HDR_CNT       = 500
	DIV_HASH_LEN          = 5
	MAX_REQ_BLK_ONCE      = 16
	UPDATE_RATE_PER_BLOCK = 2
)

const (
	HELLO_TIMEOUT      = 3 // Seconds
	MAX_HELLO_RETYR    = 3
	MAX_BUF_LEN        = 1024 * 16 // Fixme The maximum buffer to receive message
	MAX_CHAN_BUF       = 512
	PROTOCOL_VERSION   = 0
	PERIOD_UPDATE_TIME = 3 // Time to update and sync information with other nodes
	HEARTBEAT          = 2
	KEEPALIVE_TIMEOUT  = 3
	DIAL_TIMEOUT       = 6
	CONN_MONITOR       = 6
	CONN_MAX_BACK      = 4000 // ms
	MAX_RETRY_COUNT    = 3
	MAX_SYNC_HDR_REQ   = 2 //Max Concurrent Sync Header Request
)

// The node state
const (
	INIT        = 0
	HAND        = 1
	HAND_SHAKE  = 2
	HAND_SHAKED = 3
	ESTABLISH   = 4
	INACTIVITY  = 5
)

var ReceiveDuplicateBlockCnt uint64 //an index to detecting networking status

type Noder interface {
	Version() uint32
	GetID() uint64
	Services() uint64
	GetAddr() string
	GetPort() uint16
	GetHttpInfoPort() int
	SetHttpInfoPort(uint16)
	GetHttpInfoState() bool
	SetHttpInfoState(bool)
	GetState() uint32
	GetRelay() bool
	SetState(state uint32)
	GetPubKey() keypair.PublicKey
	CompareAndSetState(old, new uint32) bool
	UpdateRXTime(t time.Time)
	LocalNode() Noder
	OnDelNode(id uint64) (Noder, bool)
	OnAddNode(Noder)
	CloseConn()
	GetHeight() uint64
	GetConnectionCnt() uint
	ExistedID(id common.Uint256) bool
	ReqNeighborList()
	DumpInfo()
	UpdateInfo(t time.Time, version uint32, services uint64,
		port uint16, nonce uint64, relay uint8, height uint64)
	ConnectSeeds()
	Connect(nodeAddr string) error
	Tx(buf []byte)
	GetTime() int64
	NodeEstablished(uid uint64) bool
	GetEvent(eventName string) *events.Event
	GetNeighborAddrs() ([]NodeAddr, uint64)
	IncRxTxnCnt()
	GetTxnCnt() uint64
	GetRxTxnCnt() uint64

	Xmit(interface{}) error
	GetBookkeeperAddr() keypair.PublicKey
	GetBookkeepersAddrs() ([]keypair.PublicKey, uint64)
	SetBookkeeperAddr(pk keypair.PublicKey)
	GetNeighborHeights() ([]uint64, uint64)

	GetNeighborNoder() []Noder
	GetNbrNodeCnt() uint32
	GetLastRXTime() time.Time
	SetHeight(height uint64)
	WaitForPeersStart()
	WaitForSyncBlkFinish()
	IsAddrInNbrList(addr string) bool
	SetAddrInConnectingList(addr string) bool
	RemoveAddrInConnectingList(addr string)
	AddInRetryList(addr string)
	RemoveFromRetryList(addr string)

	OnHeaderReceive(headers []*types.Header)
	OnBlockReceive(block *types.Block)
}

func (msg *NodeAddr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg NodeAddr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}
