package node

import (
	"DNA/common"
	"DNA/common/log"
	. "DNA/config"
	"DNA/core/ledger"
	"DNA/core/transaction"
	"DNA/crypto"
	. "DNA/net/message"
	. "DNA/net/protocol"
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
	//sync.RWMutex	//The Lock not be used as expected to use function channel instead of lock
	state     uint32 // node state
	id        uint64 // The nodes's id
	cap       uint32 // The node capability set
	version   uint32 // The network protocol the node used
	services  uint64 // The services the node supplied
	relay     bool   // The relay capability of the node (merge into capbility flag)
	height    uint64 // The node latest block height
	publicKey *crypto.PubKey
	// TODO does this channel should be a buffer channel
	chF        chan func() error // Channel used to operate the node without lock
	link                         // The link status and infomation
	local      *node             // The pointer to local node
	nbrNodes                     // The neighbor node connect with currently node except itself
	eventQueue                   // The event queue to notice notice other modules
	TXNPool                      // Unconfirmed transaction pool
	idCache                      // The buffer to store the id of the items which already be processed
}

func (node node) DumpInfo() {
	fmt.Printf("Node info:\n")
	fmt.Printf("\t state = %d\n", node.state)
	fmt.Printf("\t id = 0x%x\n", node.id)
	fmt.Printf("\t addr = %s\n", node.addr)
	fmt.Printf("\t conn = %v\n", node.conn)
	fmt.Printf("\t cap = %d\n", node.cap)
	fmt.Printf("\t version = %d\n", node.version)
	fmt.Printf("\t services = %d\n", node.services)
	fmt.Printf("\t port = %d\n", node.port)
	fmt.Printf("\t relay = %v\n", node.relay)
	fmt.Printf("\t height = %v\n", node.height)
	fmt.Printf("\t conn cnt = %v\n", node.link.connCnt)
}

func (node *node) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint64, relay uint8, height uint64) {

	node.UpdateTime(t)
	node.id = nonce
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
	runtime.SetFinalizer(&n, rmNode)
	go n.backend()
	return &n
}

func InitNode(pubKey *crypto.PubKey) Noder {
	n := NewNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.link.port = uint16(Parameters.NodePort)
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	// Fixme replace with the real random number
	n.id = uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
	fmt.Printf("Init node ID to 0x%0x \n", n.id)
	n.nbrNodes.init()
	n.local = n
	n.publicKey = pubKey
	n.TXNPool.init()
	n.eventQueue.init()

	go n.initConnection()
	go n.updateNodeInfo()

	return n
}

func rmNode(node *node) {
	log.Debug(fmt.Sprintf("Remove unused/deuplicate node: 0x%0x", node.id))
}

// TODO pass pointer to method only need modify it
func (node *node) backend() {
	for f := range node.chF {
		f()
	}
}

func (node node) GetID() uint64 {
	return node.id
}

func (node node) GetState() uint32 {
	return atomic.LoadUint32(&(node.state))
}

func (node node) getConn() net.Conn {
	return node.conn
}

func (node node) GetPort() uint16 {
	return node.port
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

func (node *node) SetState(state uint32) {
	atomic.StoreUint32(&(node.state), state)
}

func (node *node) CompareAndSetState(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&(node.state), old, new)
}

func (node *node) LocalNode() Noder {
	return node.local
}

func (node node) GetHeight() uint64 {
	return node.height
}

func (node *node) UpdateTime(t time.Time) {
	node.time = t
}

func (node node) Xmit(inv common.Inventory) error {
	common.Trace()
	var buffer []byte
	var err error

	if inv.Type() == common.TRANSACTION {
		log.Info("****TX transaction message*****\n")
		transaction, ret := inv.(*transaction.Transaction)
		if ret {
			//transaction.Serialize(tmpBuffer)
			buffer, err = NewTxn(transaction)
			if err != nil {
				log.Warn("Error New Tx message ", err.Error())
				return err
			}
		}

	} else if inv.Type() == common.BLOCK {
		log.Info("****TX block message****\n")
		block, isBlock := inv.(*ledger.Block)
		// FiXME, should be moved to higher layer
		if isBlock == false {
			log.Warn("Wrong block be Xmit")
			return errors.New("Wrong block be Xmit")
		}

		err := ledger.DefaultLedger.Blockchain.AddBlock(block)
		if err != nil {
			log.Warn("Add block error")
			return errors.New("Add block error before Xmit")
		}
		buffer, err = NewBlock(block)
		if err != nil {
			log.Warn("Error New Block message ", err.Error())
			return err
		}
	} else if inv.Type() == common.CONSENSUS {
		log.Info("*****TX consensus message****\n")
		payload, ret := inv.(*ConsensusPayload)
		if ret {
			buffer, err = NewConsensus(payload)
			if err != nil {
				log.Warn("Error New consensus message ", err.Error())
				return err
			}
		}
	} else {
		log.Info("Unknown Xmit message type")
		return errors.New("Unknow Xmit message type\n")
	}

	node.nbrNodes.Broadcast(buffer)

	return nil
}

func (node node) GetAddr() string {
	return node.addr
}

func (node node) GetAddr16() ([16]byte, error) {
	var result [16]byte
	ip := net.ParseIP(node.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (node node) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

func (node node) GetMinerAddr() *crypto.PubKey {
	return node.publicKey
}

func (node node) GetMinersAddrs() ([]*crypto.PubKey, uint64) {
	pks := make([]*crypto.PubKey, 1)
	pks[0] = node.publicKey
	var i uint64
	i = 1
	//TODO read lock
	for _, n := range node.nbrNodes.List {
		if n.GetState() == ESTABLISH {
			pktmp := n.GetMinerAddr()
			pks = append(pks, pktmp)
			i++
		}
	}
	return pks, i
}

func (node *node) SetMinerAddr(pk *crypto.PubKey) {
	node.publicKey = pk
}
