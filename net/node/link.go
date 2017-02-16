package node

import (
	"GoOnchain/common"
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type link struct {
	addr  string    // The address of the node
	conn  net.Conn  // Connect socket with the peer node
	port  uint16    // The server port of the node
	time  time.Time // The latest time the node activity
	rxBuf struct {  // The RX buffer of this node to solve mutliple packets problem
		p   []byte
		len int
	}
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(node *node, buf []byte) {
	var msgLen int
	var msgBuf []byte

	if node.rxBuf.p == nil {
		if len(buf) < MSGHDRLEN {
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
		msgBuf = append(node.rxBuf.p, buf[0:msgLen]...)
		go HandleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0

		unpackNodeBuf(node, buf[msgLen:])
	}

	// TODO we need reset the node.rxBuf.p pointer and length if CheckSUM error happened?
}

func (node *node) rx() error {
	conn := node.getConn()
	from := conn.RemoteAddr().String()

	for {
		buf := make([]byte, MAXBUFLEN)
		len, err := conn.Read(buf[0:(MAXBUFLEN - 1)])
		buf[MAXBUFLEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			unpackNodeBuf(node, buf[0:len])
			//go handleNodeMsg(node, buf, len)
			break
		case io.EOF:
			//fmt.Println("Reading EOF of network conn")
			break
		default:
			fmt.Printf("read error\n", err)
			goto disconnect
		}
	}

disconnect:
	err := conn.Close()
	node.SetState(INACTIVITY)
	fmt.Printf("Close connection\n", from)
	return err
}

// Init the server port, should be run in another thread
func (n *node) initRx() {
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(NODETESTPORT))
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
		node := NewNode()
		// Currently we use the address as the ID
		node.id = conn.RemoteAddr().String()
		node.addr = conn.RemoteAddr().String()
		node.local = n
		fmt.Println("Remote node %s connect with %s\n",
			conn.RemoteAddr(), conn.LocalAddr())
		node.conn = conn
		// TOOD close the conn when erro happened
		// TODO lock the node and assign the connection to Node.
		n.neighb.add(node)
		go node.rx()
	}
	//TODO When to free the net listen resouce?
}

func (node *node) Connect(nodeAddr string) {
	node.chF <- func() error {
		common.Trace()
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			fmt.Println("Error dialing\n", err.Error())
			return err
		}

		n := NewNode()
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
		node.neighb.add(n)
		go n.rx()
		return nil
	}
}

// TODO construct a TX channel and other application just drop the message to the channel
func (node node) Tx(buf []byte) {
	node.chF <- func() error {
		common.Trace()
		_, err := node.conn.Write(buf)
		if err != nil {
			fmt.Println("Error sending messge to peer node\n", err.Error())
		}
		return err
	}
}

// func (net net) Xmit(inv Inventory) error {
// 	//if (!KnownHashes.Add(inventory.Hash)) return false;
// 	t := inv.Type()
// 	switch t {
// 	case BLOCK:
//                 if (Blockchain.Default == null) {
// 			return false
// 		}
//                 Block block = (Block)inventory;
//                 if (Blockchain.Default.ContainsBlock(block.Hash)) {
// 			return false;
// 		}
//                 if (!Blockchain.Default.AddBlock(block)) {
// 			return false;
// 		}
// 	case TRANSACTION:
// 		if (!AddTransaction((Transaction)inventory)) {
// 			return false
// 		}
// 	case CONSENSUS:
//                 if (!inventory.Verify()) {
// 			return false
// 		}
// 	default:
// 		fmt.Print("Unknow inventory type/n")
// 		return errors.New("Unknow inventory type/n")
// 	}

// 	RelayCache.Add(inventory);
// 	foreach (RemoteNode node in connectedPeers)
// 	relayed |= node.Relay(inventory);
// 	NewInventory.Invoke(this, inventory);
// }
