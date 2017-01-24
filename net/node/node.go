package node

import (
	"fmt"
	"strconv"
	"net"
	"io"
	"time"
	"errors"
	"runtime"
	"sync/atomic"
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	. "GoOnchain/net/message"
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
	local		*node		// The pointer to local node
	private		*uint		// Reserver for future using
}

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

func (node node) GetHandshakeTime() (time.Time) {
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

func (node *node) UpdateTime(t time.Time) {
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
		fmt.Printf("The Rx msg payload is %d\n", PayloadLen(buf))
		msgLen = PayloadLen(buf) + MSGHDRLEN
	} else {
		msgLen = node.rxBuf.len
	}

	//fmt.Printf("The msg length is %d, buf len is %d\n", msgLen, len(buf))
	if len(buf) == msgLen {
		msgBuf = append(node.rxBuf.p, buf[:]...)
		go HandleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0
	} else if len(buf) < msgLen {
		node.rxBuf.p = append(node.rxBuf.p, buf[:]...)
		node.rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(node.rxBuf.p, buf[0 : msgLen]...)
		go HandleNodeMsg(node, msgBuf, len(msgBuf))
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
	node.SetState(INACTIVITY)
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
		Nodes.add(node)
		go node.rx()
	}
	//TODO When to free the net listen resouce?
}

func (node *node) Connect(nodeAddr string)  {
	node.chF <- func() {
		common.Trace()
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			fmt.Println("Error dialing\n", err.Error())
			return
		}

		n := newNode()
		n.conn = conn
		n.id = conn.RemoteAddr().String()
		n.addr = conn.RemoteAddr().String()
		// FixMe Only for testing
		n.height = 1000
		n.local = node

		fmt.Printf("Connect node %s connect with %s with %s\n",
			conn.LocalAddr().String(), conn.RemoteAddr().String(),
			conn.RemoteAddr().Network())
		// TODO Need lock
		Nodes.add(n)
		go n.rx()
	}
}

// TODO construct a TX channel and other application just drop the message to the channel
func (node node) Tx(buf []byte) {
	node.chF <- func() {
		common.Trace()
		_, err := node.conn.Write(buf)
		if err != nil {
			fmt.Println("Error sending messge to peer node\n", err.Error())
		}
		return
	}
}
