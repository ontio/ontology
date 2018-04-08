package link

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//"github.com/Ontology/events"
	. "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"

	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}
type RxBuf struct {
	// The RX buffer of this node to solve mutliple packets problem
	p   []byte
	len int
}

type Link struct {
	//Todo Add lock here
	addr     string   // The address of the node
	conn     net.Conn // Connect socket with the peer node
	port     uint16   // The server port of the node
	rxBuf    RxBuf
	time     time.Time // The latest time the node activity
	connCnt  uint64    // The connection count
	recvChan chan MsgPayload
	ConnectingNodes
}

//If there is connection return true
func (link *Link) Valid() bool {
	return link.conn != nil
}

//set message channel for link layer
func (link *Link) SetChan(msgchan chan MsgPayload) {
	link.recvChan = msgchan
}

//get address
func (link *Link) GetAddr() string {
	return link.addr
}

//set port number
func (link *Link) SetPort(p uint16) {
	link.port = p
}

//get port number
func (link *Link) GetPort() uint16 {
	return link.port
}

//get connection
func (link *Link) GetConn() net.Conn {
	return link.getConn()
}

//get connection count in total
func (link *Link) GetConnCnt() uint64 {
	return link.connCnt
}
func (link *Link) getConn() net.Conn {
	return link.conn
}

//record latest getting message time
func (link *Link) UpdateRXTime(t time.Time) {
	link.time = t
}

//UpdateRXTime return the latest message time
func (link *Link) GetRXTime() time.Time {
	return link.time
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(link *Link, buf []byte) {
	var msgLen int
	var msgBuf []byte

	if len(buf) == 0 {
		return
	}

	var rxBuf *RxBuf
	rxBuf = &link.rxBuf

	if rxBuf.len == 0 {
		length := MSG_HDR_LEN - len(rxBuf.p)
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
		//go msg.HandleNodeMsg(msgBuf, len(msgBuf))
		//use channel to send p2p message
		p2pMsg := MsgPayload{
			Id:      P2PMSG,
			Payload: msgBuf,
			Len:     len(msgBuf),
		}
		link.recvChan <- p2pMsg
		rxBuf.p = nil
		rxBuf.len = 0
	} else if len(buf) < msgLen {
		rxBuf.p = append(rxBuf.p, buf[:]...)
		rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(rxBuf.p, buf[0:msgLen]...)
		//go msg.HandleNodeMsg(msgBuf, len(msgBuf))
		//use channel to send p2p message
		p2pMsg := MsgPayload{
			Id:      P2PMSG,
			Payload: msgBuf,
			Len:     len(msgBuf),
		}
		link.recvChan <- p2pMsg
		rxBuf.p = nil
		rxBuf.len = 0

		unpackNodeBuf(link, buf[msgLen:])
	}
}

func (link *Link) rx() {
	conn := link.getConn()
	buf := make([]byte, MAX_BUF_LEN)
	for {
		len, err := conn.Read(buf[0:(MAX_BUF_LEN - 1)])
		buf[MAX_BUF_LEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			t := time.Now()
			link.UpdateRXTime(t)
			unpackNodeBuf(link, buf[0:len])
		case io.EOF:
			//log.Error("Rx io.EOF: ", err, ", node id is ", node.GetID())
			goto DISCONNECT
		default:
			log.Error("Read connection error ", err)
			goto DISCONNECT
		}
	}

DISCONNECT:
	//node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeConsensusDisconnect, node)
	//use channel to send message
	disconnectMsg := MsgPayload{
		Id: DISCONNECT,
	}
	link.recvChan <- disconnectMsg
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
func (link *Link) closeConn() {
	link.conn.Close()
}

//close connection
func (link *Link) CloseConn() {
	link.closeConn()
}

//establishing the connection to remote peers and listening for incoming peers
func (link *Link) InitConnection() {
	isTls := Parameters.IsTLS
	var listener net.Listener
	var err error
	if isTls {
		listener, err = initTlsListen()
		if err != nil {
			log.Error("TLS listen failed")
			return
		}
	} else {
		listener, err = initNonTlsListen()
		if err != nil {
			log.Error("non TLS listen failed")
			return
		}
	}
	go link.waitForConnect(listener)
	//TODO Release the net listen resouce
}

func (link *Link) waitForConnect(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		link.connCnt++

		link.addr, err = parseIPaddr(conn.RemoteAddr().String())
		link.conn = conn

		go link.rx()
	}
}

func initNonTlsListen() (net.Listener, error) {
	log.Debug()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(Parameters.NodePort)))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
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

	log.Info("TLS listen port is ", strconv.Itoa(int(Parameters.NodePort)))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(int(Parameters.NodePort)), tlsConfig)
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

//record the peer which is going to be dialed and sent version message but not in establish state
func (link *Link) SetAddrInConnectingList(addr string) (added bool) {
	link.ConnectingNodes.Lock()
	defer link.ConnectingNodes.Unlock()
	for _, a := range link.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	link.ConnectingAddrs = append(link.ConnectingAddrs, addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (link *Link) RemoveAddrInConnectingList(addr string) {
	link.ConnectingNodes.Lock()
	defer link.ConnectingNodes.Unlock()
	addrs := []string{}
	for i, a := range link.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(link.ConnectingAddrs[:i], link.ConnectingAddrs[i+1:]...)
		}
	}
	link.ConnectingAddrs = addrs
}

//Connect
func (link *Link) Connect(nodeAddr string) error {
	log.Debug()

	if added := link.SetAddrInConnectingList(nodeAddr); added == false {
		return errors.New("node exist in connecting list, cancel")
	}

	isTls := Parameters.IsTLS
	var conn net.Conn
	var err error

	if isTls {
		conn, err = TLSDial(nodeAddr)
		if err != nil {
			link.RemoveAddrInConnectingList(nodeAddr)
			log.Error("TLS connect failed: ", err)
			return err
		}
	} else {
		conn, err = NonTLSDial(nodeAddr)
		if err != nil {
			link.RemoveAddrInConnectingList(nodeAddr)
			log.Error("non TLS connect failed: ", err)
			return err
		}
	}
	link.connCnt++

	link.conn = conn
	link.addr, err = parseIPaddr(conn.RemoteAddr().String())

	log.Info(fmt.Sprintf("Connect node %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))

	go link.rx()
	connectMsg := MsgPayload{
		Id: CONNECT,
	}
	link.recvChan <- connectMsg
	return nil
}

func NonTLSDial(nodeAddr string) (net.Conn, error) {
	log.Debug()
	conn, err := net.DialTimeout("tcp", nodeAddr, time.Second*DIAL_TIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//Dial with TLS
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
	dialer.Timeout = time.Second * DIAL_TIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (link *Link) Tx(buf []byte) error {
	return link.tx(buf)
}

func (link *Link) tx(buf []byte) error {
	log.Debugf("TX buf length: %d\n%x", len(buf), buf)

	_, err := link.conn.Write(buf)
	if err != nil {
		log.Error("Error sending messge to peer node ", err.Error())
		//node.local.eventQueue.GetEvent("disconnect").Notify(events.EventNodeDisconnect, node)
		//use channel to send message
		disconnectMsg := MsgPayload{
			Id: DISCONNECT,
		}
		link.recvChan <- disconnectMsg
		return err
	}

	return nil
}
