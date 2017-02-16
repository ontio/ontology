package protocol

import (
	"GoOnchain/common"
	"GoOnchain/events"
	"GoOnchain/core/transaction"
	"time"
)

const (
	MSGCMDLEN   = 12
	CMDOFFSET   = 4
	CHECKSUMLEN = 4
	HASHLEN     = 32 // hash length in byte
	MSGHDRLEN   = 24
	NETMAGIC    = 0x74746e41 // Keep the same as antshares only for testing
)
const (
	HELLOTIMEOUT  = 3 // Seconds
	MAXHELLORETYR = 3
	MAXBUFLEN     = 1024 * 1024 * 5 // Fixme The maximum buffer to receive message
	MAXCHANBUF    = 512
	//NETMAGIC	 = 0x414d5446 // Keep the same as antshares only for testing
	PROTOCOLVERSION = 0

	NODETESTPORT     = 20333 // TODO get from config file
	PERIODUPDATETIME = 3     // Time to update and sync information with other nodes
)

// The node state
const (
	INIT         = 0
	HANDSHAKEING = 1
	HANDSHAKED   = 2
	ESTABLISH    = 3
	INACTIVITY   = 4
)

type Noder interface {
	Version() uint32
	GetID() string
	Services() uint64
	GetPort() uint16
	GetState() uint
	GetNonce() uint32
	GetRelay() bool
	SetState(state uint)
	GetHandshakeTime() time.Time
	SetHandshakeTime(t time.Time)
	GetHandshakeRetry() uint32
	SetHandshakeRetry(r uint32)
	UpdateTime(t time.Time)
	LocalNode() Noder
	GetHeight() uint64
	GetConnectionCnt() uint
	GetTxnPool() map[common.Uint256]*transaction.Transaction
	AppendTxnPool(*transaction.Transaction) bool
	ExistedID(id common.Uint256) bool
	ReqNeighborList()
	//GetTxn(common.Uint256) transaction.Transaction
	Connect(nodeAddr string)
	//Xmit(inv Inventory) error // The transmit interface
	Tx(buf []byte)
}

type Tmper interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
	Connect(nodeAddr string)
}

type JsonNoder interface {
	GetConnectionCnt() uint
	GetTxnPool() map[common.Uint256]*transaction.Transaction
	Xmit(common.Inventory) error
	GetTransaction(hash common.Uint256) *transaction.Transaction	
}
