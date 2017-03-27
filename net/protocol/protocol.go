package protocol

import (
	"DNA/common"
	"DNA/core/transaction"
	"DNA/crypto"
	"DNA/events"
	"bytes"
	"encoding/binary"
	"time"
)

type NodeAddr struct {
	Time     int64
	Services uint64
	IpAddr   [16]byte
	Port     uint16
	ID       uint64 // Unique ID
}

const (
	MSGCMDLEN    = 12
	CMDOFFSET    = 4
	CHECKSUMLEN  = 4
	HASHLEN      = 32 // hash length in byte
	MSGHDRLEN    = 24
	NETMAGIC     = 0x74746e41
	MAXBLKHDRCNT = 2000
	MAXINVHDRCNT = 500
)
const (
	HELLOTIMEOUT     = 3 // Seconds
	MAXHELLORETYR    = 3
	MAXBUFLEN        = 1024 * 1024 * 5 // Fixme The maximum buffer to receive message
	MAXCHANBUF       = 512
	PROTOCOLVERSION  = 0
	PERIODUPDATETIME = 3 // Time to update and sync information with other nodes
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

type Noder interface {
	Version() uint32
	GetID() uint64
	Services() uint64
	GetPort() uint16
	GetState() uint32
	GetRelay() bool
	SetState(state uint32)
	CompareAndSetState(old, new uint32) bool
	UpdateTime(t time.Time)
	LocalNode() Noder
	DelNbrNode(id uint64) (Noder, bool)
	AddNbrNode(Noder)
	CloseConn()
	GetHeight() uint64
	GetConnectionCnt() uint
	GetTxnPool(bool) map[common.Uint256]*transaction.Transaction
	AppendTxnPool(*transaction.Transaction) bool
	ExistedID(id common.Uint256) bool
	ReqNeighborList()
	DumpInfo()
	UpdateInfo(t time.Time, version uint32, services uint64,
		port uint16, nonce uint64, relay uint8, height uint64)
	Connect(nodeAddr string) error
	Tx(buf []byte)
	GetTime() int64
	NodeEstablished(uid uint64) bool
	GetEvent(eventName string) *events.Event
	GetNeighborAddrs() ([]NodeAddr, uint64)
	GetTransaction(hash common.Uint256) *transaction.Transaction
	Xmit(common.Inventory) error
	SynchronizeTxnPool()
	GetMinerAddr() *crypto.PubKey
	GetMinersAddrs() ([]*crypto.PubKey, uint64)
	SetMinerAddr(pk *crypto.PubKey)
}

type JsonNoder interface {
	GetConnectionCnt() uint
	GetTxnPool(bool) map[common.Uint256]*transaction.Transaction
	Xmit(common.Inventory) error
	GetTransaction(hash common.Uint256) *transaction.Transaction
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
