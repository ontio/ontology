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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/msg_pack"
	"github.com/Ontology/p2pserver/peer"
)

//NetServer represent all the actions in net layer
type NetServer struct {
	Self     *peer.Peer
	SyncChan chan types.MsgPayload
	ConsChan chan types.MsgPayload
	ConnectingNodes
	PeerSyncAddress map[string]*peer.Peer
	PeerConsAddress map[string]*peer.Peer
}

type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

//InitListen start listening on the config port
func (n *NetServer) Start() {
	n.InitConnection()
}

//GetVersion return self peer`s version
func (n *NetServer) GetVersion() uint32 {
	return n.Self.GetVersion()
}

//GetPort return self peer`s txn port
func (n *NetServer) GetPort() uint16 {
	return n.Self.GetSyncPort()
}

//GetConsensusPort return self peer`s consensus port
func (n *NetServer) GetConsensusPort() uint16 {
	return n.Self.GetConsPort()
}

//GetId return peer`s id
func (n *NetServer) GetId() uint64 {
	return n.Self.GetID()
}

//GetTime return the last contact time of self peer
func (n *NetServer) GetTime() int64 {
	return n.Self.GetTimeStamp()
}

//GetState return the self peer`s state
func (n *NetServer) GetState() uint32 {
	return n.Self.GetSyncState()
}

//GetServices return the service state of self peer
func (n *NetServer) GetServices() uint64 {
	return n.Self.GetServices()
}

//GetNeighborAddrs return all the nbr peer`s addr
func (n *NetServer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return n.Self.Np.GetNeighborAddrs()
}

//GetConnectionCnt return the total number of valid connections
func (n *NetServer) GetConnectionCnt() uint32 {
	return n.Self.Np.GetNbrNodeCnt()
}

//Tx send data buf to peer
func (n *NetServer) Send(p *peer.Peer, data []byte, isConsensus bool) error {
	if p != nil {
		return p.Send(data, isConsensus)
	}
	log.Error("send to a invalid peer")
	return errors.New("send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (n *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return n.Self.Np.NodeEstablished(p.GetID())
	}
	return false

}

func (n *NetServer) Connect(addr string, isConsensus bool) error {
	log.Debug()

	if added := n.addInConnectingList(addr); added == false {
		return errors.New("node exist in connecting list, cancel")
	}

	isTls := config.Parameters.IsTLS
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			n.removeFromConnectingList(addr)
			log.Error("TLS connect failed: ", err)
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			n.removeFromConnectingList(addr)
			log.Error("non TLS connect failed: ", err)
			return err
		}
	}
	addr, err = parseIPAddr(conn.RemoteAddr().String())
	log.Info(fmt.Sprintf("Connect node %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network()))

	if !isConsensus {
		remotePeer = peer.NewPeer()
		n.PeerSyncAddress[addr] = remotePeer
		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.AttachSyncChan(n.ConsChan)
		go remotePeer.SyncLink.Rx()
		remotePeer.SetSyncState(types.HAND)
		vpl := msgpack.NewVersionPayload(n.Self, false)
		buf, _ := msgpack.NewVersion(vpl, n.Self.GetPubKey())
		remotePeer.SyncLink.Tx(buf)
	} else {
		remotePeer = peer.NewPeer() //would merge with a exist peer in versionhandle
		n.PeerConsAddress[addr] = remotePeer
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachConsChan(n.ConsChan)
		go remotePeer.ConsLink.Rx()
		remotePeer.SetConsState(types.HAND)
		vpl := msgpack.NewVersionPayload(n.Self, true)
		buf, _ := msgpack.NewVersion(vpl, n.Self.GetPubKey())
		remotePeer.ConsLink.Tx(buf)
	}

	return nil
}

//Halt stop all net layer logic
func (n *NetServer) Halt() {
	peers := n.Self.Np.GetNeighbors()
	for _, p := range peers {
		p.CloseSync()
		p.CloseCons()
	}
	n.Self.CloseSync()
	n.Self.CloseCons()
}

//establishing the connection to remote peers and listening for incoming peers
func (n *NetServer) InitConnection() error {
	isTls := config.Parameters.IsTLS
	var synclistener net.Listener
	var conslistener net.Listener
	var err error

	syncPort := n.Self.SyncLink.GetPort()
	consPort := n.Self.ConsLink.GetPort()

	if syncPort == 0 {
		log.Error("Sync Port invalid")
		return errors.New("Sync Port invalid")
	}
	if isTls {
		synclistener, err = initTlsListen(syncPort)
		if err != nil {
			log.Error("TLS listen failed")
			return errors.New("Sync TLS listen failed")
		}
	} else {
		synclistener, err = initNonTlsListen(syncPort)
		if err != nil {
			log.Error("Sync non TLS listen failed")
			return errors.New("Sync non TLS listen failed")
		}
	}
	go n.startSyncAccept(synclistener)
	log.Infof("Start listen on sync port %d", syncPort)

	//consensus
	if consPort == 0 {
		//still work
		log.Error("Consensus Port invalid")
	} else {
		if isTls {
			conslistener, err = initTlsListen(consPort)
			if err != nil {
				log.Error("TLS listen failed")
				return errors.New("Sync TLS listen failed")
			}
		} else {
			conslistener, err = initNonTlsListen(consPort)
			if err != nil {
				log.Error("Sync non TLS listen failed")
				return errors.New("Sync non TLS listen failed")
			}
		}
		go n.startConsAccept(conslistener)
		log.Infof("Start listen on sync port %d", syncPort)
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
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr, err := parseIPAddr(conn.RemoteAddr().String())
		n.PeerSyncAddress[addr] = remotePeer
		if err != nil {
			log.Errorf("Error parse remote ip:%s", conn.RemoteAddr().String())
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
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		remotePeer := peer.NewPeer()
		addr, err := parseIPAddr(conn.RemoteAddr().String())
		n.PeerConsAddress[addr] = remotePeer
		if err != nil {
			log.Errorf("Error parse remote ip:%s", conn.RemoteAddr().String())
			return
		}
		remotePeer.ConsLink.SetAddr(addr)
		remotePeer.ConsLink.SetConn(conn)
		remotePeer.AttachSyncChan(n.ConsChan)
		go remotePeer.ConsLink.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (n *NetServer) addInConnectingList(addr string) (added bool) {
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
func (n *NetServer) removeFromConnectingList(addr string) {
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

func initNonTlsListen(port uint16) (net.Listener, error) {
	log.Debug()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

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

func parseIPAddr(s string) (string, error) {
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("Split IP address&port error")
		return s, errors.New("Split IP address&port error")
	}
	return s[:i], nil
}

func nonTLSDial(addr string) (net.Conn, error) {
	log.Debug()
	conn, err := net.DialTimeout("tcp", addr, time.Second*types.DIAL_TIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//Dial with TLS
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
	dialer.Timeout = time.Second * types.DIAL_TIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
