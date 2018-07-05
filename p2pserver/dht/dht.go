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

package dht

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"bytes"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"strconv"
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
}

// NewDHT returns an instance of DHT with the given id
func NewDHT(id types.NodeID) *DHT {
	dht := &DHT{
		nodeID:         id,
		addr:           config.DefConfig.Genesis.DHT.IP,
		udpPort:        uint16(config.DefConfig.Genesis.DHT.UDPPort),
		tcpPort:        uint16(config.DefConfig.P2PNode.NodePort),
		routingTable:   &routingTable{},
		bootstrapNodes: make(map[types.NodeID]*types.Node, 0),
	}

	dht.init()
	return dht
}

func loadSeeds() []*types.Node {
	seeds := make([]*types.Node, 0, len(config.DefConfig.Genesis.DHT.Seeds))
	for i := 0; i < len(config.DefConfig.Genesis.DHT.Seeds); i++ {
		node := config.DefConfig.Genesis.DHT.Seeds[i]
		seed := &types.Node{
			IP:      node.IP,
			UDPPort: node.UDPPort,
			TCPPort: node.TCPPort,
		}
		log.Infof("loadSeeds: ip %s port %d", seed.IP, seed.UDPPort)
		id := types.ConstructID(seed.IP, seed.UDPPort)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, id)
		copy(seed.ID[:], b[:])
		seeds = append(seeds, seed)
	}
	return seeds
}

func (this *DHT) SetPort(tcpPort uint16, udpPort uint16) {
	this.tcpPort = tcpPort
	this.udpPort = udpPort
}

// init initializes an instance of DHT
func (this *DHT) init() {
	this.recvCh = make(chan *types.DHTMessage, types.MSG_CACHE)
	this.stopCh = make(chan struct{})
	this.messagePool = types.NewRequestPool(this.onRequestTimeOut)
	this.feedCh = make(chan *types.FeedEvent, types.MSG_CACHE)
	this.routingTable.init(this.nodeID, this.feedCh)
}

// Start starts DHT service
func (this *DHT) Start() {
	seeds := loadSeeds()
	for _, seed := range seeds {
		this.bootstrapNodes[seed.ID] = seed
	}

	go this.loop()

	err := this.listenUDP(":" + strconv.Itoa(int(this.udpPort)))
	if err != nil {
		log.Errorf("listen udp failed.")
	}
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
	this.DisplayRoutingTable()

	log.Info("start lookup")
	this.lookup(this.nodeID)
}

// add node to routing table in synchronize
func (this *DHT) syncAddNodes(nodes map[types.NodeID]*types.Node) {
	waitRequestIds := make([]types.RequestId, 0)
	for _, n := range nodes {
		addr, err := getNodeUDPAddr(n)
		if err != nil {
			log.Infof("node %s address is error!", n.ID)
			continue
		}
		requestId, isNewRequest := this.messagePool.AddRequest(n,
			types.DHT_PING_REQUEST, nil, true)
		if isNewRequest {
			this.ping(addr)
		}
		waitRequestIds = append(waitRequestIds, requestId)
	}
	this.messagePool.Wait(waitRequestIds)
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
			return
		case <-refresh.C:
			go this.refreshRoutingTable()
		}
	}
}

// refreshRoutingTable refreshs k bucket
func (this *DHT) refreshRoutingTable() {
	log.Info("refreshRoutingTable start")
	// Todo:
	this.syncAddNodes(this.bootstrapNodes)
	results := this.lookup(this.nodeID)
	if results != nil && len(results) > 0 {
		return
	}

	var targetID types.NodeID
	rand.Read(targetID[:])
	log.Infof("refreshRoutingTable: %s", targetID.String())
	this.lookup(targetID)
}

// lookup executes a network search for nodes closest to the given
// target and setup k bucket
func (this *DHT) lookup(targetID types.NodeID) []*types.Node {
	bucket, _ := this.routingTable.locateBucket(targetID)
	node, ret := this.routingTable.isNodeInBucket(targetID, bucket)
	if ret == true {
		log.Infof("targetID %s is in the bucket %d", targetID.String(), bucket)
		return []*types.Node{node}
	}

	closestNodes := this.routingTable.getClosestNodes(types.BUCKET_SIZE, targetID)
	if len(closestNodes) == 0 {
		return nil
	}

	visited := make(map[types.NodeID]bool)
	knownNode := make(map[types.NodeID]bool)
	pendingQueries := 0

	visited[this.nodeID] = true

	for {
		for i := 0; i < len(closestNodes) && pendingQueries < types.FACTOR; i++ {
			node := closestNodes[i]
			if visited[node.ID] == true {
				continue
			}
			visited[node.ID] = true
			pendingQueries++
			go func() {
				this.findNode(node, targetID)
				this.messagePool.AddRequest(node, types.DHT_FIND_NODE_REQUEST, nil, false)
			}()
		}

		if pendingQueries == 0 {
			break
		}

		this.waitAndHandleResponse(knownNode, closestNodes, targetID)
		pendingQueries--
	}
	return closestNodes
}

// waitAndHandleResponse waits for the result
func (this *DHT) waitAndHandleResponse(knownNode map[types.NodeID]bool, closestNodes []*types.Node, targetID types.NodeID) {
	responseCh := this.messagePool.GetResultChan()
	select {
	case entries, ok := <-responseCh:
		if ok {
			for _, n := range entries {
				// Todo:
				if knownNode[n.ID] == true || n.ID == this.nodeID {
					continue
				}
				knownNode[n.ID] = true
				if len(closestNodes) < types.BUCKET_SIZE {
					closestNodes = append(closestNodes, n)
				} else {
					index := len(closestNodes)
					for i, entry := range closestNodes {
						for j := range targetID {
							da := entry.ID[j] ^ targetID[j]
							db := n.ID[j] ^ targetID[j]
							if da > db {
								index = i
								break
							}
						}
					}

					if index < len(closestNodes) {
						closestNodes[index] = n
					}
				}
			}
		}
	}

}

// onRequestTimeOut handles a timeout event of request
func (this *DHT) onRequestTimeOut(requestId types.RequestId) {
	reqType := types.GetReqTypeFromReqId(requestId)
	this.messagePool.DeleteRequest(requestId)
	log.Info("request ", requestId, "timeout!")
	if reqType == types.DHT_FIND_NODE_REQUEST {
		results := make([]*types.Node, 0)
		this.messagePool.SetResults(results)
	} else if reqType == types.DHT_PING_REQUEST {
		pendingNode, ok := this.messagePool.GetSupportData(requestId)
		if ok && pendingNode != nil {
			bucketIndex, _ := this.routingTable.locateBucket(pendingNode.ID)
			this.routingTable.addNode(pendingNode, bucketIndex)
		}
	}
}

// findNode sends findNode to remote node to get the closest nodes to target
func (this *DHT) findNode(remotePeer *types.Node, targetID types.NodeID) error {
	addr, err := getNodeUDPAddr(remotePeer)
	if err != nil {
		return err
	}
	findNodeMsg := msgpack.NewFindNode(this.nodeID, targetID)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, findNodeMsg)
	this.send(addr, bf.Bytes())
	return nil
}

// addNode adds a node to the K bucket.
// remotePeer: added node
// shouldWait: if ping the lastNode located in the same k bucket of remotePeer, the request should be wait or not
func (this *DHT) addNode(remotePeer *types.Node) {
	if remotePeer == nil || remotePeer.ID == this.nodeID {
		return
	}

	// find node in own bucket
	bucketIndex, _ := this.routingTable.locateBucket(remotePeer.ID)
	remoteNode, isInBucket := this.routingTable.isNodeInBucket(remotePeer.ID, bucketIndex)
	// update peer info in local bucket
	remoteNode = remotePeer
	if isInBucket {
		this.routingTable.addNode(remoteNode, bucketIndex)
	} else {
		bucketNodeNum := this.routingTable.getTotalNodeNumInBukcet(bucketIndex)
		if bucketNodeNum < types.BUCKET_SIZE { // bucket is not full
			this.routingTable.addNode(remoteNode, bucketIndex)
		} else {
			lastNode := this.routingTable.getLastNodeInBucket(bucketIndex)
			addr, err := getNodeUDPAddr(lastNode)
			if err != nil {
				this.routingTable.removeNode(lastNode.ID)
				this.routingTable.addNode(remoteNode, bucketIndex)
				return
			}
			if _, isNewRequest := this.messagePool.AddRequest(lastNode,
				types.DHT_PING_REQUEST, remotePeer, false); isNewRequest {
				this.ping(addr)
			}
		}
	}
	return
}

// ping the remote node
func (this *DHT) ping(addr *net.UDPAddr) error {
	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	pingMsg := msgpack.NewDHTPing(this.nodeID, this.udpPort, this.tcpPort, ip, addr)
	bf := new(bytes.Buffer)
	err := mt.WriteMessage(bf, pingMsg)
	if err != nil {
		log.Info(err)
		return err
	}
	this.send(addr, bf.Bytes())
	return nil
}

// pong reply remote node when receiving ping
func (this *DHT) pong(addr *net.UDPAddr) error {

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}

	pongMsg := msgpack.NewDHTPong(this.nodeID, this.udpPort, this.tcpPort, ip, addr)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, pongMsg)
	this.send(addr, bf.Bytes())
	return nil
}

// findNodeReply reply remote node when receiving find node
func (this *DHT) findNodeReply(addr *net.UDPAddr, targetId types.NodeID) error {
	// query routing table
	nodes := this.routingTable.getClosestNodes(types.BUCKET_SIZE, targetId)

	neighborsMsg := msgpack.NewNeighbors(this.nodeID, nodes)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, neighborsMsg)
	this.send(addr, bf.Bytes())

	return nil
}

// processPacket invokes the related handler to process the packet
func (this *DHT) processPacket(from *net.UDPAddr, packet []byte) {
	msg, err := mt.ReadMessage(bytes.NewBuffer(packet))
	if err != nil {
		log.Info("receive dht message error:", err)
		return
	}
	msgType := msg.CmdType()
	log.Infof("Recv UDP msg %s %v", msgType, from)
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
		log.Infof("processPacket: unknown msg %s", msgType)
	}
}

// recvUDPMsg waits for the udp msg and puts it to the msg queue
func (this *DHT) recvUDPMsg() {
	defer this.conn.Close()
	buf := make([]byte, common.MAX_BUF_LEN)
	for {
		nbytes, from, err := this.conn.ReadFromUDP(buf)
		if err != nil {
			log.Error("ReadFromUDP error:", err)
			return
		}
		// Todo:
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
		log.Error("failed to resolve udp address", laddr, "error: ", err)
		return err
	}
	this.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Error("failed to listen udp on", addr, "error: ", err)
		return err
	}
	log.Infof("DHT is listening on %s", laddr)
	go this.recvUDPMsg()
	return nil
}

// send a msg to the remote node
func (this *DHT) send(addr *net.UDPAddr, msg []byte) error {
	_, err := this.conn.WriteToUDP(msg, addr)
	if err != nil {
		log.Error("failed to send msg", err)
		return err
	}
	return nil
}

func getNodeUDPAddr(node *types.Node) (*net.UDPAddr, error) {
	addr := new(net.UDPAddr)
	addr.IP = net.ParseIP(node.IP).To16()
	if addr.IP == nil {
		log.Error("Parse IP address error\n", node.IP)
		return nil, errors.New("Parse IP address error")
	}
	addr.Port = int(node.UDPPort)
	return addr, nil
}

func (this *DHT) DisplayRoutingTable() {
	for bucketIndex, bucket := range this.routingTable.buckets {
		if this.routingTable.getTotalNodeNumInBukcet(bucketIndex) == 0 {
			continue
		}
		fmt.Println("[", bucketIndex, "]: ")
		for i := 0; i < this.routingTable.getTotalNodeNumInBukcet(bucketIndex); i++ {
			fmt.Printf("%x %s %d %d\n", bucket.entries[i].ID[:], bucket.entries[i].IP,
				bucket.entries[i].UDPPort, bucket.entries[i].TCPPort)
		}
	}
}
