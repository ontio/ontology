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

// Package common privides DHT/Kad protocol
package dht

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

const (
	DHT_BLACK_LIST_FILE = "./dht_black_list"
	DHT_WHITE_LIST_FILE = "./dht_white_list"
)

// DHT manage the DHT/Kad protocol resource, mainly including
// route table, the channel to netserver, the udp message queue
type DHT struct {
	mu             sync.Mutex
	version        uint16                       // Local DHT version
	nodeID         types.NodeID                 // Local DHT id
	routingTable   *routingTable                // The k buckets
	addr           string                       // Local Address
	udpPort        uint16                       // Local UDP port
	tcpPort        uint16                       // Local TCP port
	conn           *net.UDPConn                 // UDP listen fd
	messagePool    *types.DHTMessagePool        // Manage the request msgs(ping, findNode)
	recvCh         chan *types.DHTMessage       // The queue to receive msg from UDP network
	bootstrapNodes map[types.NodeID]*types.Node // Hold inital nodes from configure and peer file to contact
	feedCh         chan *types.FeedEvent        // Notify netserver of add/del a remote peer
	stopCh         chan struct{}                // Stop DHT module

	whiteList []string
	blackList []string
}

// NewDHT returns an instance of DHT with the given id
func NewDHT(id types.NodeID) *DHT {
	dht := &DHT{
		nodeID:         id,
		addr:           config.DefConfig.P2PNode.NetworkMgrCfg.DHT.IP,
		udpPort:        uint16(config.DefConfig.P2PNode.NetworkMgrCfg.DHT.UDPPort),
		tcpPort:        uint16(config.DefConfig.P2PNode.NodePort),
		routingTable:   &routingTable{},
		bootstrapNodes: make(map[types.NodeID]*types.Node, 0),
	}

	dht.init()
	return dht
}

// loadSeeds load seed nodes as initial nodes to contact
func loadSeeds() []*types.Node {
	seeds := make([]*types.Node, 0, len(config.DefConfig.P2PNode.NetworkMgrCfg.DHT.Seeds))
	for i := 0; i < len(config.DefConfig.P2PNode.NetworkMgrCfg.DHT.Seeds); i++ {
		node := config.DefConfig.P2PNode.NetworkMgrCfg.DHT.Seeds[i]
		ns, err := net.LookupHost(node.IP)
		if err != nil {
			log.Warnf("resolve ip %s err %s", node.IP, err.Error())
			continue
		}

		seed := &types.Node{
			IP:      ns[0],
			UDPPort: node.UDPPort,
			TCPPort: node.TCPPort,
		}
		id := types.ConstructID(seed.IP, seed.UDPPort)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, id)
		copy(seed.ID[:], b[:])
		seeds = append(seeds, seed)
	}
	return seeds
}

// init initializes an instance of DHT
func (this *DHT) init() {
	this.recvCh = make(chan *types.DHTMessage, types.MSG_CACHE)
	this.stopCh = make(chan struct{})
	this.messagePool = types.NewRequestPool(this.onRequestTimeOut)
	this.feedCh = make(chan *types.FeedEvent, types.MSG_CACHE)
	this.routingTable.init(this.nodeID, this.feedCh)

	// load white list and black list
	this.loadWhiteList()
	this.loadBlackList(DHT_BLACK_LIST_FILE)
}

// Start starts DHT service
func (this *DHT) Start() {
	seeds := loadSeeds()
	for _, seed := range seeds {
		if seed.ID != this.nodeID {
			this.bootstrapNodes[seed.ID] = seed
		}
	}
	err := this.listenUDP(":" + strconv.Itoa(int(this.udpPort)))
	if err != nil {
		log.Errorf("listen udp failed.")
		return
	}

	go this.loop()

	this.bootstrap()
}

// Stop stops DHT service
func (this *DHT) Stop() {
	if this.stopCh != nil {
		this.stopCh <- struct{}{}
	}

	if this.feedCh != nil {
		close(this.feedCh)
	}
	// close udp connect
	this.conn.Close()
}

//SetFallbackNodes appends recent connected peers
func (this *DHT) SetFallbackNodes(nodes []types.Node) {
	for _, n := range nodes {
		this.bootstrapNodes[n.ID] = &n
	}
}

// bootstrap loads initial node and setup k bucket
func (this *DHT) bootstrap() {
	// Todo:
	this.syncAddNodes(this.bootstrapNodes)

	log.Debug("DHT starts lookup")
	this.lookup(this.nodeID)
}

// add node to routing table in synchronize
func (this *DHT) syncAddNodes(nodes map[types.NodeID]*types.Node) {
	waitGroup := new(sync.WaitGroup)
	for _, node := range nodes {
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			log.Debugf("[dht]node %s address is error!", node.ID)
			continue
		}
		_, isNewRequest := this.messagePool.AddRequest(node,
			types.DHT_PING_REQUEST, nil, waitGroup)
		if isNewRequest {
			this.ping(addr)
		}
	}
	waitGroup.Wait()
}

// GetFeecCh returns the feed event channel
func (this *DHT) GetFeedCh() chan *types.FeedEvent {
	return this.feedCh
}

// loop runs the periodical process
func (this *DHT) loop() {
	refresh := time.NewTicker(types.REFRESH_INTERVAL)
	for {
		select {
		case pk, ok := <-this.recvCh:
			if ok {
				go this.processPacket(pk.From, pk.Payload)
			}
		case <-this.stopCh:
			refresh.Stop()
			return
		case <-refresh.C:
			go this.refreshRoutingTable()
		}
	}
}

// refreshRoutingTable refresh k bucket
func (this *DHT) refreshRoutingTable() {
	log.Debug("[dht]refreshRoutingTable start, add bootstrapNodes")
	// Todo:
	this.syncAddNodes(this.bootstrapNodes)
	this.lookup(this.nodeID)
	var targetID types.NodeID
	rand.Read(targetID[:])
	log.Debugf("[dht]refreshRoutingTable: target id %s", targetID.String())
	this.lookup(targetID)
}

// lookup executes a network search for nodes closest to the given
// target and setup k bucket
func (this *DHT) lookup(targetID types.NodeID) types.ClosestList {
	node, _ := this.routingTable.queryNode(targetID)
	if node != nil {
		closestList := make(types.ClosestList, 1)
		item := &types.Item{
			Entry:    node,
			Distance: getDistance(targetID, node.ID),
		}
		closestList[0] = item
		return closestList
	}

	closestList := this.routingTable.getClosestNodes(types.BUCKET_SIZE, targetID)
	if closestList.Len() == 0 {
		return nil
	}

	visited := make(map[types.NodeID]bool)
	knownNode := make(map[types.NodeID]bool)
	pendingQueries := 0

	visited[this.nodeID] = true

	for {
		for i := 0; i < closestList.Len() && pendingQueries < types.FACTOR; i++ {
			item := closestList[i]
			if visited[item.Entry.ID] == true {
				continue
			}
			visited[item.Entry.ID] = true
			pendingQueries++
			go func() {
				this.findNode(item.Entry, targetID)
				this.messagePool.AddRequest(item.Entry, types.DHT_FIND_NODE_REQUEST, nil, nil)
			}()
		}

		if pendingQueries == 0 {
			break
		}

		this.waitAndHandleResponse(knownNode, closestList, targetID)
		pendingQueries--
	}
	return closestList
}

// waitAndHandleResponse waits for the result
func (this *DHT) waitAndHandleResponse(knownNode map[types.NodeID]bool, closestList types.ClosestList,
	targetID types.NodeID) {
	responseCh := this.messagePool.GetResultChan()
	entries, ok := <-responseCh
	if ok {
		for _, n := range entries {
			// Todo:
			if knownNode[n.ID] == true || n.ID == this.nodeID {
				continue
			}
			knownNode[n.ID] = true
			push(n, targetID, closestList, types.BUCKET_SIZE)
		}
	}

}

// addNode adds a node to the K bucket.
// remotePeer: added node
// shouldWait: if ping the lastNode located in the same k bucket of remotePeer, the request should be wait or not
func (this *DHT) addNode(remotePeer *types.Node) {
	if remotePeer == nil || remotePeer.ID == this.nodeID {
		return
	}

	// find node in own bucket
	remoteNode, bucketIndex := this.routingTable.queryNode(remotePeer.ID)

	if remoteNode != nil {
		// update peer info in local bucket
		remoteNode = remotePeer
		this.routingTable.addNode(remoteNode, bucketIndex)
	} else {
		// add new peer in local bucket
		remoteNode = remotePeer
		bucketNodeNum := this.routingTable.getTotalNodeNumInBukcet(bucketIndex)
		if bucketNodeNum < types.BUCKET_SIZE { // bucket is not full
			this.routingTable.addNode(remoteNode, bucketIndex)
		} else {
			lastNode := this.routingTable.getLastNodeInBucket(bucketIndex)
			addr, err := getNodeUDPAddr(lastNode)
			if err != nil {
				log.Debugf("[dht]addnode: node ip %s, udp %d, tcp %d", remoteNode.IP, remoteNode.UDPPort, remoteNode.TCPPort)
				this.routingTable.removeNode(lastNode.ID)
				this.routingTable.addNode(remoteNode, bucketIndex)
				return
			}
			if _, isNewRequest := this.messagePool.AddRequest(lastNode,
				types.DHT_PING_REQUEST, remoteNode, nil); isNewRequest {
				this.ping(addr)
			}
		}
	}
	return
}

// processPacket invokes the related handler to process the packet
func (this *DHT) processPacket(from *net.UDPAddr, packet []byte) {
	msg, _, err := mt.ReadMessage(bytes.NewBuffer(packet))
	if err != nil {
		log.Debugf("[dht]processPacket: receive dht message error: %v", err)
		return
	}
	msgType := msg.CmdType()
	log.Debugf("[dht]processPacket: UDP msg %s from %v", msgType, from)
	switch msgType {
	case common.DHT_PING:
		this.pingHandle(from, msg)
	case common.DHT_PONG:
		this.pongHandle(from, msg)
	case common.DHT_FIND_NODE:
		this.findNodeHandle(from, msg)
	case common.DHT_NEIGHBORS:
		this.neighborsHandle(from, msg)
	default:
		log.Debugf("[dht]processPacket: unknown msg %s", msgType)
	}
}

// recvUDPMsg waits for the udp msg and puts it to the msg queue
func (this *DHT) recvUDPMsg() {
	defer this.conn.Close()
	buf := make([]byte, common.MAX_BUF_LEN)
	for {
		nbytes, from, err := this.conn.ReadFromUDP(buf)
		if err != nil {
			log.Errorf("recvUDPMsg: ReadFromUDP error: %v", err)
			return
		}

		if this.isInBlackList(from.IP.String()) {
			log.Debugf("recvUDPMsg: receive a msg from %v in blacklist",
				from)
			continue
		}

		if !this.isInWhiteList(from.IP.String()) {
			log.Debugf("recvUDPMsg: receive a msg from %v is not in whitelist",
				from)
			continue
		}

		pk := &types.DHTMessage{
			From:    from,
			Payload: make([]byte, 0, nbytes),
		}
		pk.Payload = append(pk.Payload, buf[:nbytes]...)
		this.recvCh <- pk
	}
}

// listenUDP listens on the specified address:port
func (this *DHT) listenUDP(laddr string) error {
	addr, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		log.Errorf("listenUDP: failed to resolve udp address %s, err %v",
			laddr, err)
		return err
	}
	this.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("listenUDP: failed to listen udp on %s err %v", addr, err)
		return err
	}
	log.Debugf("DHT is listening on %s", laddr)
	go this.recvUDPMsg()
	return nil
}

// send a msg to the remote node
func (this *DHT) send(addr *net.UDPAddr, msg mt.Message) error {
	sink := comm.NewZeroCopySink(nil)
	mt.WriteMessage(sink, msg)
	_, err := this.conn.WriteToUDP(sink.Bytes(), addr)
	if err != nil {
		log.Errorf("DHT failed to send msg to addr %v. err %v", addr, err)
		return err
	}
	return nil
}

// AddBlackList adds a node to blacklist
func (this *DHT) AddBlackList(addr string) {
	this.blackList = append(this.blackList, addr)
}

// AddWhiteList adds a node to whitelist
func (this *DHT) AddWhiteList(addr string) {
	this.whiteList = append(this.whiteList, addr)
}

// SaveBlackListToFile saves blacklist to a local file
func (this *DHT) SaveBlackListToFile() {
	this.saveListToFile(this.blackList, DHT_BLACK_LIST_FILE)
}

// SaveWhiteListToFile saves whitelist to a local file
func (this *DHT) SaveWhiteListToFile() {
	this.saveListToFile(this.whiteList, DHT_WHITE_LIST_FILE)
}

// getNodeUDPAddr returns UDP address with a given node
func getNodeUDPAddr(node *types.Node) (*net.UDPAddr, error) {
	addr := new(net.UDPAddr)
	addr.IP = net.ParseIP(node.IP).To16()
	if addr.IP == nil {
		log.Errorf("getNodeUDPAddr: failed to parse IP %s", node.IP)
		return nil, errors.New("Parse IP address error")
	}
	addr.Port = int(node.UDPPort)
	return addr, nil
}

// Resolve searches for a specific node with the given ID.
func (this *DHT) Resolve(id uint64) types.ClosestList {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID types.NodeID
	copy(nodeID[:], b[:])

	node, _ := this.routingTable.queryNode(nodeID)
	if node != nil {
		closestList := make(types.ClosestList, 1)
		item := &types.Item{
			Entry:    node,
			Distance: getDistance(nodeID, node.ID),
		}
		closestList[0] = item
		return closestList
	}

	closestList := this.routingTable.getClosestNodes(types.FACTOR, nodeID)
	return closestList
}
