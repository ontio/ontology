package protocol

import (
	"DNA/common"
	"DNA/core/ledger"
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
	MSGCMDLEN     = 12
	CMDOFFSET     = 4
	CHECKSUMLEN   = 4
	HASHLEN       = 32 // hash length in byte
	MSGHDRLEN     = 24
	NETMAGIC      = 0x74746e41
	MAXBLKHDRCNT  = 2000
	MAXINVHDRCNT  = 500
	DIVHASHLEN    = 5
	MINCONNCNT    = 3
	MAXREQBLKONCE = 16
)
const (
	HELLOTIMEOUT     = 3 // Seconds
	MAXHELLORETYR    = 3
	MAXBUFLEN        = 1024 * 1024 * 5 // Fixme The maximum buffer to receive message
	MAXCHANBUF       = 512
	PROTOCOLVERSION  = 0
	PERIODUPDATETIME = 3 // Time to update and sync information with other nodes
	HEARTBEAT        = 2
	KEEPALIVETIMEOUT = 3
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
	UpdateRXTime(t time.Time)
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
	ConnectSeeds()
	Connect(nodeAddr string) error
	Tx(buf []byte)
	GetTime() int64
	NodeEstablished(uid uint64) bool
	GetEvent(eventName string) *events.Event
	GetNeighborAddrs() ([]NodeAddr, uint64)
	GetTransaction(hash common.Uint256) *transaction.Transaction
	IncRxTxnCnt()
	GetTxnCnt() uint64
	GetRxTxnCnt() uint64

	Xmit(interface{}) error
	SynchronizeTxnPool()
	GetBookKeeperAddr() *crypto.PubKey
	GetBookKeepersAddrs() ([]*crypto.PubKey, uint64)
	SetBookKeeperAddr(pk *crypto.PubKey)
	GetNeighborHeights() ([]uint64, uint64)
	SyncNodeHeight()
	CleanSubmittedTransactions(block *ledger.Block) error

	IsSyncHeaders() bool
	SetSyncHeaders(b bool)
	IsSyncFailed() bool
	SetSyncFailed()
	StartRetryTimer()
	StopRetryTimer()
	GetNeighborNoder() []Noder
	GetNbrNodeCnt() uint32
	StoreFlightHeight(height uint32)
	GetFlightHeightCnt() int
	RemoveFlightHeightLessThan(height uint32)
	RemoveFlightHeight(height uint32)
	GetLastRXTime() time.Time
	SetHeight(height uint64)
	WaitForFourPeersStart()
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
