/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package netserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

//NewNetServer return the net object in p2p
func NewNetServer(pubKey keypair.PublicKey) p2p.P2P {

	n := &NetServer{
		SyncChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
		ConsChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
	}

	n.PeerAddrMap.PeerSyncAddress = make(map[string]*peer.Peer)
	n.PeerAddrMap.PeerConsAddress = make(map[string]*peer.Peer)

	n.init(pubKey)
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base         peer.PeerCom
	synclistener net.Listener
	conslistener net.Listener
	SyncChan     chan *common.MsgPayload
	ConsChan     chan *common.MsgPayload
	ConnectingNodes
	PeerAddrMap
	Np *peer.NbrPeers
}

//ConnectingNodes include all addr in connecting state
type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerSyncAddress map[string]*peer.Peer
	PeerConsAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (n *NetServer) init(pubKey keypair.PublicKey) error {
	n.base.SetVersion(common.PROTOCOL_VERSION)
	if config.Parameters.NodeType == common.SERVICE_NODE_NAME {
		n.base.SetServices(uint64(common.SERVICE_NODE))
	} else if config.Parameters.NodeType == common.VERIFY_NODE_NAME {
		n.base.SetServices(uint64(common.VERIFY_NODE))
	}

	if config.Parameters.NodeConsensusPort == 0 || config.Parameters.NodePort == 0 ||
		config.Parameters.NodeConsensusPort == config.Parameters.NodePort {
		log.Error("Network port invalid, please check config.json")
		return errors.New("Invalid port")
	}
	n.base.SetSyncPort(config.Parameters.NodePort)
	n.base.SetConsPort(config.Parameters.NodeConsensusPort)

	n.base.SetRelay(true)

	var id uint64
	key := keypair.SerializePublicKey(pubKey)
	err := binary.Read(bytes.NewBuffer(key[:8]), binary.LittleEndian, &(id))
	if err != nil {
		log.Error(err)
		return err
	}
	n.base.SetID(id)

	log.Info(fmt.Sprintf("Init peer ID to 0x%x", n.base.GetID()))
	n.Np = &peer.NbrPeers{}
	n.Np.Init()

	n.base.SetPubKey(pubKey)
	return nil
}

//InitListen start listening on the config port
func (n *NetServer) Start() {
	n.InitConnection()
}

//GetVersion return self peer`s version
func (n *NetServer) GetVersion() uint32 {
	return n.base.GetVersion()
}

//GetId return peer`s id
func (n *NetServer) GetID() uint64 {
	return n.base.GetID()
}

// SetHeight sets the local's height
func (n *NetServer) SetHeight(height uint64) {
	n.base.SetHeight(height)
}

// GetHeight return peer's heigh
func (n *NetServer) GetHeight() uint64 {
	return n.base.GetHeight()
}

//GetTime return the last contact time of self peer
func (n *NetServer) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

//GetServices return the service state of self peer
func (n *NetServer) GetServices() uint64 {
	return n.base.GetServices()
}

//GetSyncPort return the sync port
func (n *NetServer) GetSyncPort() uint16 {
	return n.base.GetSyncPort()
}

//GetConsPort return the cons port
func (n *NetServer) GetConsPort() uint16 {
	return n.base.GetConsPort()
}

//GetHttpInfoPort return the port support info via http
func (n *NetServer) GetHttpInfoPort() uint16 {
	return n.base.GetHttpInfoPort()
}

//GetRelay return whether net module can relay msg
func (n *NetServer) GetRelay() bool {
	return n.base.GetRelay()
}

//GetPubKey return the key config in net module
func (n *NetServer) GetPubKey() keypair.PublicKey {
	return n.base.GetPubKey()
}

// GetPeer returns a peer with the peer id
func (n *NetServer) GetPeer(id uint64) *peer.Peer {
	return n.Np.GetPeer(id)
}

//return nbr peers collection
func (n *NetServer) GetNp() *peer.NbrPeers {
	return n.Np
}

//GetNeighborAddrs return all the nbr peer`s addr
func (n *NetServer) GetNeighborAddrs() ([]common.PeerAddr, uint64) {
	return n.Np.GetNeighborAddrs()
}

//GetConnectionCnt return the total number of valid connections
func (n *NetServer) GetConnectionCnt() uint32 {
	return n.Np.GetNbrNodeCnt()
}

//AddNbrNode add peer to nbr peer list
func (n *NetServer) AddNbrNode(remotePeer *peer.Peer) {
	n.Np.AddNbrNode(remotePeer)
}

//DelNbrNode delete nbr peer by id
func (n *NetServer) DelNbrNode(id uint64) (*peer.Peer, bool) {
	return n.Np.DelNbrNode(id)
}

//GetNeighbors return all nbr peer
func (n *NetServer) GetNeighbors() []*peer.Peer {
	return n.Np.GetNeighbors()
}

//NodeEstablished return whether a peer is establish with self according to id
func (n *NetServer) NodeEstablished(id uint64) bool {
	return n.Np.NodeEstablished(id)
}

//Xmit called by actor, broadcast msg
func (n *NetServer) Xmit(buf []byte, isCons bool) {
	n.Np.Broadcast(buf, isCons)
}

//GetMsgChan return sync or consensus channel when msgrouter need msg input
func (n *NetServer) GetMsgChan(isConsensus bool) chan *common.MsgPayload {
	if isConsensus {
		return n.ConsChan
	} else {
		return n.SyncChan
	}
}

//Tx send data buf to peer
func (n *NetServer) Send(p *peer.Peer, data []byte, isConsensus bool) error {
	if p != nil {
		if config.Parameters.DualPortSurpport == false {
			return p.Send(data, false)
		}
		return p.Send(data, isConsensus)
	}
	log.Error("send to a invalid peer")
	return errors.New("send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (n *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return n.Np.NodeEstablished(p.GetID())
	}
	return false

}

//Connect used to connect net address under sync or cons mode
func (n *NetServer) Connect(addr string, isConsensus bool) error {
	if n.IsNbrPeerAddr(addr, isConsensus) {
		return nil
	}
	if added := n.AddInConnectingList(addr); added == false {
		log.Info("node exist in connecting list", addr)
		return errors.New("node exist in connecting list")
	}

	isTls := config.Parameters.IsTLS
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			n.RemoveFromConnectingList(addr)
			log.Error("connect failed: ", err)
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			n.RemoveFromConnectingList(addr)
			log.Error("connect failed: ", err)
			return err
		}
	}

	addr = conn.RemoteAddr().String()
	log.Info(fmt.Sprintf("peer %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))

	if !isConsensus {
		remotePeer = peer.NewPeer()
		n.AddPeerSyncAddress(addr, remotePeer)
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(n.SyncChan)
		go remotePeer.SyncLink.Rx()
		remotePeer.SetSyncState(common.HAND)
		vpl := msgpack.NewVersionPayload(n, false)
		buf, _ := msgpack.NewVersion(vpl, n.GetPubKey())
		remotePeer.SyncLink.Tx(buf)
	} else {
		remotePeer = peer.NewPeer() //would merge with a exist peer in versionhandle
		n.AddPeerConsAddress(addr, remotePeer)
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(n.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetConsState(common.HAND)
		vpl := msgpack.NewVersionPayload(n, true)
		buf, _ := msgpack.NewVersion(vpl, n.GetPubKey())
		remotePeer.ConsLink.Tx(buf)
	}

	return nil
}

//Halt stop all net layer logic
func (n *NetServer) Halt() {
	peers := n.Np.GetNeighbors()
	for _, p := range peers {
		p.CloseSync()
		p.CloseCons()
	}
	if n.synclistener != nil {
		n.synclistener.Close()
	}
	if n.conslistener != nil {
		n.conslistener.Close()
	}

}

//establishing the connection to remote peers and listening for incoming peers
func (n *NetServer) InitConnection() error {
	isTls := config.Parameters.IsTLS

	var err error

	syncPort := n.base.GetSyncPort()
	consPort := n.base.GetConsPort()

	if syncPort == 0 {
		log.Error("Sync Port invalid")
		return errors.New("Sync Port invalid")
	}
	if isTls {
		n.synclistener, err = initTlsListen(syncPort)
		if err != nil {
			log.Error("Sync listen failed")
			return errors.New("Sync listen failed")
		}
	} else {
		n.synclistener, err = initNonTlsListen(syncPort)
		if err != nil {
			log.Error("Sync listen failed")
			return errors.New("Sync listen failed")
		}
	}
	go n.startSyncAccept(n.synclistener)
	log.Infof("Start listen on sync port %d", syncPort)

	//consensus
	if config.Parameters.DualPortSurpport == false {
		log.Info("Dual port mode not supported,keep single link")
		return nil
	}
	if consPort == 0 || consPort == syncPort {
		//still work
		log.Error("Consensus Port invalid,keep single link")
	} else {
		if isTls {
			n.conslistener, err = initTlsListen(consPort)
			if err != nil {
				log.Error("Cons listen failed")
				return errors.New("Cons listen failed")
			}
		} else {
			n.conslistener, err = initNonTlsListen(consPort)
			if err != nil {
				log.Error("Cons listen failed")
				return errors.New("Cons listen failed")
			}
		}
		go n.startConsAccept(n.conslistener)
		log.Infof("Start listen on consensus port %d", consPort)
	}
	return nil
}

//startAccept start listen to sync port
func (n *NetServer) startSyncAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote sync node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		n.AddPeerSyncAddress(addr, remotePeer)
		if err != nil {
			log.Errorf("Error parse remote ip:%s", addr)
			return
		}
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(n.SyncChan)
		go remotePeer.SyncLink.Rx()
	}
}

//startAccept start listen to Consensus port
func (n *NetServer) startConsAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote cons node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		n.AddPeerConsAddress(addr, remotePeer)
		if err != nil {
			log.Errorf("Error parse remote ip:%s", addr)
			return
		}
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachSyncChan(n.ConsChan)
		go remotePeer.ConsLink.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (n *NetServer) AddInConnectingList(addr string) (added bool) {
	n.ConnectingNodes.Lock()
	defer n.ConnectingNodes.Unlock()
	for _, a := range n.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	n.ConnectingAddrs = append(n.ConnectingAddrs, addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (n *NetServer) RemoveFromConnectingList(addr string) {
	n.ConnectingNodes.Lock()
	defer n.ConnectingNodes.Unlock()
	addrs := []string{}
	for i, a := range n.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			addrs = append(n.ConnectingAddrs[:i], n.ConnectingAddrs[i+1:]...)
		}
	}
	n.ConnectingAddrs = addrs
}

//find exist peer from addr map
func (n *NetServer) GetPeerFromAddr(addr string) *peer.Peer {
	var p *peer.Peer
	n.PeerAddrMap.Lock()
	defer n.PeerAddrMap.Unlock()

	p, ok := n.PeerSyncAddress[addr]
	if ok {
		return p
	}
	p, ok = n.PeerConsAddress[addr]
	if ok {
		return p
	}
	return nil
}

//initNonTlsListen return net.Listener with nonTls mode
func initNonTlsListen(port uint16) (net.Listener, error) {
	log.Debug()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

//initTlsListen return net.Listener with Tls mode
func initTlsListen(port uint16) (net.Listener, error) {
	CertPath := config.Parameters.CertPath
	KeyPath := config.Parameters.KeyPath
	CAPath := config.Parameters.CAPath

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

	log.Info("TLS listen port is ", strconv.Itoa(int(port)))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(int(port)), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}

//nonTLSDial return net.Conn with nonTls
func nonTLSDial(addr string) (net.Conn, error) {
	log.Debug()
	conn, err := net.DialTimeout("tcp", addr, time.Second*common.DIAL_TIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//TLSDial return net.Conn with TLS
func TLSDial(nodeAddr string) (net.Conn, error) {
	CertPath := config.Parameters.CertPath
	KeyPath := config.Parameters.KeyPath
	CAPath := config.Parameters.CAPath

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
	dialer.Timeout = time.Second * common.DIAL_TIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//IsNbrPeerAddr return result whether the address is under connecting
func (n *NetServer) IsNbrPeerAddr(addr string, isConsensus bool) bool {
	var addrNew string
	n.Np.RLock()
	defer n.Np.RUnlock()
	for _, p := range n.Np.List {
		if p.GetSyncState() == common.HAND || p.GetSyncState() == common.HAND_SHAKE ||
			p.GetSyncState() == common.ESTABLISH {
			if isConsensus {
				addrNew = p.ConsLink.GetAddr()
			} else {
				addrNew = p.SyncLink.GetAddr()
			}
			if strings.Compare(addrNew, addr) == 0 {
				return true
			}
		}
	}
	return false
}

//AddPeerSyncAddress add sync addr to peer-addr map
func (n *NetServer) AddPeerSyncAddress(addr string, p *peer.Peer) {
	n.PeerAddrMap.Lock()
	defer n.PeerAddrMap.Unlock()
	n.PeerSyncAddress[addr] = p
}

//AddPeerConsAddress add cons addr to peer-addr map
func (n *NetServer) AddPeerConsAddress(addr string, p *peer.Peer) {
	n.PeerAddrMap.Lock()
	defer n.PeerAddrMap.Unlock()
	n.PeerConsAddress[addr] = p
}

//RemovePeerSyncAddress remove sync addr from peer-addr map
func (n *NetServer) RemovePeerSyncAddress(addr string) {
	n.PeerAddrMap.Lock()
	defer n.PeerAddrMap.Unlock()
	if _, ok := n.PeerSyncAddress[addr]; ok {
		n.PeerSyncAddress[addr] = nil
	}
}

//RemovePeerConsAddress remove cons addr from peer-addr map
func (n *NetServer) RemovePeerConsAddress(addr string) {
	n.PeerAddrMap.Lock()
	defer n.PeerAddrMap.Unlock()
	if _, ok := n.PeerConsAddress[addr]; ok {
		n.PeerConsAddress[addr] = nil
	}
}
