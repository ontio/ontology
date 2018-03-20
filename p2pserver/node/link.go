package node

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/events"
	msg "github.com/Ontology/p2pserver/message"
	. "github.com/Ontology/p2pserver/protocol"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type RxBuf struct {
	// The RX buffer of this node to solve mutliple packets problem
	p   []byte
	len int
}

type ConsensusLink struct {
	consensusPort  uint16
	consensusConn  net.Conn // Connect socket with the peer node
	consensusRxBuf RxBuf
}

type link struct {
	//Todo Add lock here
	addr         string   // The address of the node
	conn         net.Conn // Connect socket with the peer node
	port         uint16   // The server port of the node
	rxBuf        RxBuf
	httpInfoPort uint16    // The node information server port of the node
	time         time.Time // The latest time the node activity
	connCnt      uint64    // The connection count
	//ConsensusLink
	consensusPort  uint16
	consensusConn  net.Conn // Connect socket with the peer node
	consensusRxBuf RxBuf
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(node *node, buf []byte, isConsensusChannel bool) {
	var msgLen int
	var msgBuf []byte

	if len(buf) == 0 {
		return
	}

	var rxBuf *RxBuf
	if isConsensusChannel {
		rxBuf = &node.consensusRxBuf
	} else {
		rxBuf = &node.rxBuf
	}

	if rxBuf.len == 0 {
		length := MSGHDRLEN - len(rxBuf.p)
		if length > len(buf) {
			length = len(buf)
			rxBuf.p = append(rxBuf.p, buf[0:length]...)
			return
		}

		rxBuf.p = append(rxBuf.p, buf[0:length]...)
		if msg.ValidMsgHdr(rxBuf.p) == false {
			rxBuf.p = nil
			rxBuf.len = 0
			log.Warn("Get error message header, TODO: relocate the msg header")
			// TODO Relocate the message header
			return
		}

		rxBuf.len = msg.PayloadLen(rxBuf.p)
		buf = buf[length:]
	}

	msgLen = rxBuf.len
	if len(buf) == msgLen {
		msgBuf = append(rxBuf.p, buf[:]...)
		go msg.HandleNodeMsg(node, msgBuf, len(msgBuf))
		rxBuf.p = nil
		rxBuf.len = 0
	} else if len(buf) < msgLen {
		rxBuf.p = append(rxBuf.p, buf[:]...)
		rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(rxBuf.p, buf[0:msgLen]...)
		go msg.HandleNodeMsg(node, msgBuf, len(msgBuf))
		rxBuf.p = nil
		rxBuf.len = 0

		unpackNodeBuf(node, buf[msgLen:], isConsensusChannel)
	}
}

func (node *node) rx(isConsensusChannel bool) {
	conn := node.getconn(isConsensusChannel)
	buf := make([]byte, MAXBUFLEN)
	for {
		len, err := conn.Read(buf[0:(MAXBUFLEN - 1)])
		buf[MAXBUFLEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			if !isConsensusChannel {
				t := time.Now()
				node.UpdateRXTime(t)
			}
			unpackNodeBuf(node, buf[0:len], isConsensusChannel)
		case io.EOF:
			log.Error("Rx io.EOF: ", err, ", node id is ", node.GetID())
			goto DISCONNECT
		default:
			log.Error("Read connection error ", err)
			goto DISCONNECT
		}
	}

DISCONNECT:
	if isConsensusChannel {
		node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeConsensusDisconnect, node)
	} else {
		node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeDisconnect, node)
	}
}

func printIPAddr() {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			log.Info("IPv4: ", ipv4)
		}
	}
}

func (link *link) closeConn(isConsensusChannel bool) {
	if !isConsensusChannel {
		link.conn.Close()
	} else {
		link.consensusConn.Close()
	}
}

func (link *link) CloseConn() {
	link.closeConn(false)
}

func (link *link) CloseConsensusConn() {
	link.closeConn(true)
}

func (n *node) initConnection() {
	isTls := Parameters.IsTLS
	var listener, listenerConsensus net.Listener
	var err error
	if isTls {
		listener, err = initTlsListen()
		if err != nil {
			log.Error("TLS listen failed")
			return
		}
	} else {
		listener, listenerConsensus, err = initNonTlsListen()
		if err != nil {
			log.Error("non TLS listen failed")
			return
		}
	}
	go n.waitForConnect(listener, false)
	go n.waitForConnect(listenerConsensus, true)
	//TODO Release the net listen resouce
}

func (n *node) waitForConnect(listener net.Listener, isConsensusChannel bool) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		n.link.connCnt++

		var node *node
		if !isConsensusChannel {
			node = NewNode()
			node.addr, err = parseIPaddr(conn.RemoteAddr().String())
			node.local = n
			node.conn = conn
			node.ConsensusNode = node
		} else {
			//localIp, err := parseIPaddr(conn.LocalAddr().String())
			//if err != nil {
			//	log.Error("parseIPaddr error:", err)
			//	conn.Close()
			//	continue
			//}
			//node = n.GetNbrNodeByAddr(localIp)
			//if node == nil {
			//	conn.Close()
			//	continue
			//}
			//node.consensusConn = conn
			node = NewNode()
			node.addr, err = parseIPaddr(conn.RemoteAddr().String())
			node.local = n
			node.consensusConn = conn
			node.ConsensusNode = node
		}
		go node.rx(isConsensusChannel)
	}

}

func initNonTlsListen() (net.Listener, net.Listener, error) {
	log.Debug()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, nil, err
	}
	listenerConsensus, err := net.Listen("tcp", ":"+strconv.Itoa(Parameters.NodeConsensusPort))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, nil, err
	}
	return listener, listenerConsensus, nil
}

func initTlsListen() (net.Listener, error) {
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	log.Info("TLS listen port is ", strconv.Itoa(Parameters.NodePort))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}

func parseIPaddr(s string) (string, error) {
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("Split IP address&port error")
		return s, errors.New("Split IP address&port error")
	}
	return s[:i], nil
}

func (n1 *node) Connect(nodeAddr string, isConsensusChannel bool) error {
	log.Debug()

	n := n1.LocalNode().(*node)

	//TODO consensusChannel judgement
	if !isConsensusChannel {
		if n.IsAddrInNbrList(nodeAddr) == true {
			return nil
		}
		if added := n.SetAddrInConnectingList(nodeAddr); added == false {
			return errors.New("node exist in connecting list, cancel")
		}
	}

	isTls := Parameters.IsTLS
	var conn net.Conn
	var err error

	if isTls {
		conn, err = TLSDial(nodeAddr)
		if err != nil {
			n.RemoveAddrInConnectingList(nodeAddr)
			log.Error("TLS connect failed: ", err)
			return err
		}
	} else {
		conn, err = NonTLSDial(nodeAddr)
		if err != nil {
			n.RemoveAddrInConnectingList(nodeAddr)
			log.Error("non TLS connect failed: ", err)
			return err
		}
	}
	n.link.connCnt++
	var nbrNode *node
	if isConsensusChannel {
		//TODO localnode is being or not
		n1.consensusConn = conn
		nbrNode = n1
	} else {
		nbrNode = NewNode()
		nbrNode.conn = conn
		nbrNode.addr, err = parseIPaddr(conn.RemoteAddr().String())
		nbrNode.local = n
		nbrNode.ConsensusNode = nbrNode
	}

	log.Info(fmt.Sprintf("Connect node %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))

	go nbrNode.rx(isConsensusChannel)

	if isConsensusChannel {
		nbrNode.SetConsensusState(HAND)
	} else {
		nbrNode.SetState(HAND)
	}
	buf, _ := msg.NewVersion(n, isConsensusChannel)

	nbrNode.tx(buf, isConsensusChannel)

	return nil
}

func NonTLSDial(nodeAddr string) (net.Conn, error) {
	log.Debug()
	conn, err := net.DialTimeout("tcp", nodeAddr, time.Second*DIALTIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TLSDial(nodeAddr string) (net.Conn, error) {
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		return nil, err
	}

	ret := clientCertPool.AppendCertsFromPEM(cacert)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
	}

	var dialer net.Dialer
	dialer.Timeout = time.Second * DIALTIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (node *node) Tx(buf []byte) {
	node.ConsensusNode.tx(buf, false)
}

func (node *node) ConsensusTx(buf []byte) {
	node.tx(buf, true)
}

func (node *node) tx(buf []byte, isConsensusChannel bool) {
	log.Debugf("TX buf length: %d\n%x", len(buf), buf)

	if node.GetState() == INACTIVITY {
		return
	}
	if isConsensusChannel {
		_, err := node.consensusConn.Write(buf)
		if err != nil {
			log.Error("Error sending messge to peer node ", err.Error())
			node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeConsensusDisconnect, node)
		}
	} else {
		_, err := node.conn.Write(buf)
		if err != nil {
			log.Error("Error sending messge to peer node ", err.Error())
			node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeDisconnect, node)
		}
	}
}
