package node

import (
	"GoOnchain/common"
	"GoOnchain/core/ledger"
	"GoOnchain/core/transaction"
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

// The node capability flag
const (
	RELAY        = 0x01
	SERVER       = 0x02
	NODESERVICES = 0x01
)

type node struct {
	state          uint      // node status
	id             string    // The nodes's id, MAC or IP?
	addr           string    // The address of the node
	conn           net.Conn  // Connect socket with the peer node
	nonce          uint32    // Random number to identify different entity from the same IP
	cap            uint32    // The node capability set
	version        uint32    // The network protocol the node used
	services       uint64    // The services the node supplied
	port           uint16    // The server port of the node
	relay          bool      // The relay capability of the node (merge into capbility flag)
	handshakeRetry uint32    // Handshake retry times
	handshakeTime  time.Time // Last Handshake trigger time
	height         uint64    // The node latest block height
	time           time.Time // The latest time the node activity
	// TODO does this channel should be a buffer channel
	chF   chan func() error // Channel used to operate the node without lock
	rxBuf struct {          // The RX buffer of this node to solve mutliple packets problem
		p   []byte
		len int
	}
	link           // The link status and infomation
	local  *node   // The pointer to local node
	neighb nodeMap // The neighbor node connect with currently node except itself
	//neighborNodes	*nodeMAP	// The node connect with it except the local node
	eventQueue                // The event queue to notice notice other modules
	TXNPool                   // Unconfirmed transaction pool
	idCache                   // The buffer to store the id of the items which already be processed
	ledger     *ledger.Ledger // The Local ledger
	private    *uint          // Reserver for future using
}

func (node node) DumpInfo() {
	fmt.Printf("Node info:\n")
	fmt.Printf("\t state = %d\n", node.state)
	fmt.Printf("\t id = %s\n", node.id)
	fmt.Printf("\t addr = %s\n", node.addr)
	fmt.Printf("\t conn = %v\n", node.conn)
	fmt.Printf("\t nonce = %d\n", node.nonce)
	fmt.Printf("\t cap = %d\n", node.cap)
	fmt.Printf("\t version = %d\n", node.version)
	fmt.Printf("\t services = %d\n", node.services)
	fmt.Printf("\t port = %d\n", node.port)
	fmt.Printf("\t relay = %v\n", node.relay)
	fmt.Printf("\t height = %v\n", node.height)
}

func (node *node) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint32, relay uint8, height uint32) {
	// TODO need lock
	node.UpdateTime(t)
	node.nonce = nonce
	node.version = version
	node.services = services
	node.port = port
	if relay == 0 {
		node.relay = false
	} else {
		node.relay = true
	}
	node.height = uint64(height)
}

func NewNode() *node {
	n := node{
		state: INIT,
		chF:   make(chan func() error),
	}
	// Update nonce
	runtime.SetFinalizer(&n, rmNode)
	go n.backend()
	return &n
}

func InitNode() Tmper {
	var err error
	n := NewNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.port = NODETESTPORT
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	// Fixme replace with the real random number
	n.nonce = rand.Uint32()

	n.neighb.init()
	n.local = n
	n.TXNPool.init()
	n.eventQueue.init()
	n.ledger, err = ledger.GetDefaultLedger()
	if err != nil {
		errors.New("Get Default Ledger error")
		// FIXME report the error
	}

	go n.initConnection()
	go n.updateNodeInfo()
	return n
}

func rmNode(node *node) {
	fmt.Printf("Remove node %s\n", node.addr)
}

// TODO pass pointer to method only need modify it
func (node *node) backend() {
	common.Trace()
	for f := range node.chF {
		f()
	}
}

func (node node) GetID() string {
	return node.id
}

func (node node) GetState() uint {
	return node.state
}

func (node node) getConn() net.Conn {
	return node.conn
}

func (node node) GetPort() uint16 {
	return node.port
}

func (node node) GetNonce() uint32 {
	return node.nonce
}

func (node node) GetRelay() bool {
	return node.relay
}

func (node node) Version() uint32 {
	return node.version
}

func (node node) Services() uint64 {
	return node.services
}

func (node *node) SetState(state uint) {
	node.state = state
}

func (node node) GetHandshakeTime() time.Time {
	return node.handshakeTime
}

func (node *node) SetHandshakeTime(t time.Time) {
	node.handshakeTime = t
}

func (node *node) LocalNode() Noder {
	return node.local
}

func (node node) GetHandshakeRetry() uint32 {
	return atomic.LoadUint32(&(node.handshakeRetry))
}

func (node *node) SetHandshakeRetry(r uint32) {
	node.handshakeRetry = r
	atomic.StoreUint32(&(node.handshakeRetry), r)
}

func (node node) GetHeight() uint64 {
	return node.height
}

func (node node) GetLedger() *ledger.Ledger {
	return node.ledger
}

func (node *node) UpdateTime(t time.Time) {
	node.time = t
}

func (node node) GetMemoryPool() map[common.Uint256]*transaction.Transaction {
	return node.GetTxnPool()
	// TODO refresh the pending transaction pool
}

func (node node) SynchronizeMemoryPool() {
	// Fixme need lock
	for _, n := range node.neighb.List {
		if n.state == ESTABLISH {
			ReqMemoryPool(&node)
		}
	}
}

func (node node) Xmit(inv common.Inventory) error {
	// Fixme here we only consider 1 inventory case
	var msg Inv
	t := "inv"
	copy(msg.Hdr.CMD[0:len(t)], t)
	msg.P.InvType = uint8(inv.Type())
	// FIXME filling the inventory header
	hash := inv.Hash()
	msg.P.Blk = hash[:]
	buf, _ := msg.Serialization()
	node.neighb.Broadcast(buf)
	// FIXME currenly we have no error check
	return nil
}

func (node node) GetAddr() string {
	return node.addr
}

func (node node) GetAddress() ([16]byte, error) {
	common.Trace()
	var result [16]byte
	ip := net.ParseIP(node.addr).To16()
	if (ip == nil) {
		fmt.Printf("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (node node) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

func (node node) getNbrNum() uint {
	var i uint
	for _, n := range node.local.neighb.List {
		if n.GetState() == ESTABLISH {
			i++
		}
	}
	return i
}

func (node node) GetNeighborAddrs() ([]NodeAddr, uint64) {
	var i uint64

	cnt := node.getNbrNum()
	addrs := make([]NodeAddr, cnt)
	// TODO read lock
	for _, n := range node.local.neighb.List {
		if n.GetState() == ESTABLISH {
			addrs[i].IpAddr, _ = n.GetAddress()
			addrs[i].Time = n.GetTime()
			addrs[i].Services = n.Services()
			addrs[i].Port = n.GetPort()

			i++
		}
	}
	return addrs, i
}
