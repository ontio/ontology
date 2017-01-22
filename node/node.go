package node

import (
	"fmt"
	"strconv"
	"net"
	"io"
	"sync"
	"time"
	"errors"
	"runtime"
	"math/rand"
	"sync/atomic"
	"GoOnchain/common"
)

// The node state
const (
	INIT = 0
	HANDSHAKEING = 1
	HANDSHAKED = 2
	ESTABLISH = 3
	INACTIVITY = 4
)

// The node capability flag
const (
	RELAY  = 0x01
	SERVER = 0x02
	NODESERVICES = 0x01
)

type node struct {
	state		uint		// node status
	id		string		// The nodes's id, MAC or IP?
	addr		string 		// The address of the node
	conn		net.Conn	// Connect socket with the peer node
	nonce		uint32		// Random number to identify different entity from the same IP
	cap		uint32  	// The node capability set
	version		uint32		// The network protocol the node used
	services	uint64		// The services the node supplied
	port		uint16		// The server port of the node
	relay		bool		// The relay capability of the node (merge into capbility flag)
	handshakeRetry  uint32		// Handshake retry times
	handshakeTime	time.Time	// Last Handshake trigger time
	height		uint64		// The node latest block height
	time		time.Time	// The latest time the node activity
	// TODO does this channel should be a buffer channel
	chF		chan func()	// Channel used to operate the node without lock
	rxBuf struct {			// The RX buffer of this node to solve mutliple packets problem
		p   []byte
		len int
	}
	private		*uint		// Reserver for future using
}

type nodeMap struct {
	node *node
	lock sync.RWMutex
	list map[string]*node
}

var nodes nodeMap

func newNode() (*node) {
	node := node{
		state: INIT,
		chF: make(chan func()),
	}

	// Update nonce
	runtime.SetFinalizer(&node, rmNode)
	go node.backend()
	return &node
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

func (node node) getID() string {
	return node.id
}

func (node node) getState() uint {
	return node.state
}

func (node node) getConn() net.Conn {
	return node.conn
}

func (node *node) setState(state uint) {
	node.state = state
}

func (node node) getHandshakeTime() (time.Time) {
	return node.handshakeTime
}

func (node *node) setHandshakeTime(t time.Time) {
	node.handshakeTime = t
}

func (node node) getHandshakeRetry() uint32 {
	return atomic.LoadUint32(&(node.handshakeRetry))
}

func (node *node) setHandshakeRetry(r uint32) {
	node.handshakeRetry = r
	atomic.StoreUint32(&(node.handshakeRetry), r)
}

func (node node) getHeight() uint64 {
	return node.height
}

func (node *node) updateTime(t time.Time) {
	node.time = t
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(node *node, buf []byte) {
	var msgLen int
	var msgBuf []byte

	if (node.rxBuf.p == nil) {
		if (len(buf) < MSGHDRLEN) {
			fmt.Println("Unexpected size of received message")
			errors.New("Unexpected size of received message")
			return
		}
		// FIXME Check the payload < 0 error case
		fmt.Printf("The Rx msg payload is %d\n", payloadLen(buf))
		msgLen = payloadLen(buf) + MSGHDRLEN
	} else {
		msgLen = node.rxBuf.len
	}

	//fmt.Printf("The msg length is %d, buf len is %d\n", msgLen, len(buf))
	if len(buf) == msgLen {
		msgBuf = append(node.rxBuf.p, buf[:]...)
		go handleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0
	} else if len(buf) < msgLen {
		node.rxBuf.p = append(node.rxBuf.p, buf[:]...)
		node.rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(node.rxBuf.p, buf[0 : msgLen]...)
		go handleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0

		unpackNodeBuf(node, buf[msgLen : ])
	}

	// TODO we need reset the node.rxBuf.p pointer and length if CheckSUM error happened?
}

func (node *node) rx() error {
	conn := node.getConn()
	from := conn.RemoteAddr().String()

	for {
		buf := make([]byte, MAXBUFLEN)
		len, err := conn.Read(buf[0:(MAXBUFLEN - 1)])
		buf[MAXBUFLEN - 1] = 0 //Prevent overflow
		switch err {
		case nil:
			unpackNodeBuf(node, buf[0 : len])
			//go handleNodeMsg(node, buf, len)
			break
		case io.EOF:
			//fmt.Println("Reading EOF of network conn")
			break
		default:
			fmt.Printf("read error\n", err)
			goto DISCONNECT
		}
	}

DISCONNECT:
	err := conn.Close()
	node.setState(INACTIVITY)
	fmt.Printf("Close connection\n", from)
	return err
}

// Init the server port, should be run in another thread
func (node *node) initRx () {
	listener, err := net.Listen("tcp", "localhost:" + strconv.Itoa(NODETESTPORT))
	if err != nil {
		fmt.Println("Error listening\n", err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting\n", err.Error())
			return
		}
		node := newNode()
		// Currently we use the address as the ID
		node.id = conn.RemoteAddr().String()
		node.addr = conn.RemoteAddr().String()
		fmt.Println("Remote node %s connect with %s\n",
			conn.RemoteAddr(), conn.LocalAddr())
		node.conn = conn
		// TOOD close the conn when erro happened
		// TODO lock the node and assign the connection to Node.
		nodes.add(node)
		go node.rx()
	}
	//TODO When to free the net listen resouce?
}

func (node *node) connect(nodeAddr string)  {
	node.chF <- func() {
		common.Trace()
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			fmt.Println("Error dialing\n", err.Error())
			return
		}

		node := newNode()
		node.conn = conn
		node.id = conn.RemoteAddr().String()
		node.addr = conn.RemoteAddr().String()
		// FixMe Only for testing
		node.height = 1000

		fmt.Printf("Connect node %s connect with %s with %s\n",
			conn.LocalAddr().String(), conn.RemoteAddr().String(),
			conn.RemoteAddr().Network())
		// TODO Need lock
		nodes.add(node)
		go node.rx()
	}
}

// TODO construct a TX channel and other application just drop the message to the channel
func (node node) tx(buf []byte) {
	node.chF <- func() {
		common.Trace()
		_, err := node.conn.Write(buf)
		if err != nil {
			fmt.Println("Error sending messge to peer node\n", err.Error())
		}
		return
	}
}

func (nodes *nodeMap) broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range nodes.list {
		if node.state == ESTABLISH {
			go node.tx(buf)
		}
	}
}

func (nodes *nodeMap) add(node *node) {
	//TODO lock the node Map
	// TODO check whether the node existed or not
	// TODO dupicate IP address nodes issue
	nodes.list[node.id] = node
	// Unlock the map
}

func (nodes *nodeMap) delNode(node *node) {
	//TODO lock the node Map
	delete(nodes.list, node.id)
	// Unlock the map
}

func InitNodes() {
	// TODO write lock
	n := newNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.port = NODETESTPORT
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	n.nonce = rand.Uint32()

	nodes.node = n
	nodes.list = make(map[string]*node)
}

func Relay(msgType string, msg interface{}) {
	// TODO Unicast or broadcast the message based on the type
	//node.tx()
}
