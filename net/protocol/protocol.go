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

	"github.com/Ontology/common"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	"time"
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
	VERIFYNODE  = 1
	SERVICENODE = 2
)

const (
	VERIFYNODENAME  = "verify"
	SERVICENODENAME = "service"
)

const (
	MSGCMDLEN         = 12
	CMDOFFSET         = 4
	CHECKSUMLEN       = 4
	HASHLEN           = 32 // hash length in byte
	MSGHDRLEN         = 24
	NETMAGIC          = 0x74746e41
	MAXBLKHDRCNT      = 500
	MAXINVHDRCNT      = 500
	DIVHASHLEN        = 5
	MAXREQBLKONCE     = 16
	TIMESOFUPDATETIME = 2
)

const (
	HELLOTIMEOUT     = 3 // Seconds
	MAXHELLORETYR    = 3
	MAXBUFLEN        = 1024 * 16 // Fixme The maximum buffer to receive message
	MAXCHANBUF       = 512
	PROTOCOLVERSION  = 0
	PERIODUPDATETIME = 3 // Time to update and sync information with other nodes
	HEARTBEAT        = 2
	KEEPALIVETIMEOUT = 3
	DIALTIMEOUT      = 6
	CONNMONITOR      = 6
	CONNMAXBACK      = 4000
	MAXRETRYCOUNT    = 3
	MAXSYNCHDRREQ    = 2 //Max Concurrent Sync Header Request
)

// The node state
const (
	INIT       = 0
	HAND       = 1
	HANDSHAKE  = 2
	HANDSHAKED = 3
	ESTABLISH  = 4
	INACTIVITY = 5
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
	GetPubKey() *crypto.PubKey
	CompareAndSetState(old, new uint32) bool
	UpdateRXTime(t time.Time)
	LocalNode() Noder
	DelNbrNode(id uint64) (Noder, bool)
	AddNbrNode(Noder)
	CloseConn()
	GetHeight() uint64
	GetConnectionCnt() uint
	//GetTxnPool(bool) (map[common.Uint256]*types.Transaction, common.Fixed64)
	//AppendTxnPool(*types.Transaction) ErrCode
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
	//GetTransaction(hash common.Uint256) *types.Transaction
	IncRxTxnCnt()
	GetTxnCnt() uint64
	GetRxTxnCnt() uint64

	Xmit(interface{}) error
	GetBookkeeperAddr() *crypto.PubKey
	GetBookkeepersAddrs() ([]*crypto.PubKey, uint64)
	SetBookkeeperAddr(pk *crypto.PubKey)
	GetNeighborHeights() ([]uint64, uint64)
	SyncNodeHeight()
	//CleanTransactions(txns []*types.Transaction) error

	GetNeighborNoder() []Noder
	GetNbrNodeCnt() uint32
	StoreFlightHeight(height uint32)
	GetFlightHeightCnt() int
	RemoveFlightHeightLessThan(height uint32)
	RemoveFlightHeight(height uint32)
	GetLastRXTime() time.Time
	SetHeight(height uint64)
	WaitForPeersStart()
	WaitForSyncBlkFinish()
	GetFlightHeights() []uint32
	IsAddrInNbrList(addr string) bool
	SetAddrInConnectingList(addr string) bool
	RemoveAddrInConnectingList(addr string)
	AddInRetryList(addr string)
	RemoveFromRetryList(addr string)
	AcqSyncReqSem()
	RelSyncReqSem()
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
