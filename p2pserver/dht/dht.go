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
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

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
	mu           sync.Mutex
	version      uint16                 // Local DHT version
	nodeID       types.NodeID           // Local DHT id
	routingTable *routingTable          // The k buckets
	addr         string                 // Local Address
	udpPort      uint16                 // Local UDP port
	tcpPort      uint16                 // Local TCP port
	conn         *net.UDPConn           // UDP listen fd
	messagePool  *types.DHTMessagePool  // Manage the request msgs(ping, findNode)
	recvCh       chan *types.DHTMessage // The queue to receive msg from UDP network
	seeds        []*types.Node          // Hold seed nodes from configure
	feedCh       chan *types.FeedEvent  // Notify netserver of add/del a remote peer
	stopCh       chan struct{}          // Stop DHT module
}

// NewDHT returns an instance of DHT with the given id and seed nodes
func NewDHT(id types.NodeID, seeds []*types.Node) *DHT {
	// Todo:
	dht := &DHT{
		nodeID:       id,
		addr:         "127.0.0.1",
		udpPort:      uint16(config.DefConfig.Genesis.DHT.UDPPort),
		tcpPort:      uint16(config.DefConfig.P2PNode.NodePort),
		routingTable: &routingTable{},
		seeds:        make([]*types.Node, 0, len(seeds)),
	}
	for _, seed := range seeds {
		dht.seeds = append(dht.seeds, seed)
	}
	dht.init()
	return dht
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

// bootstrap loads seed node and setup k bucket
func (this *DHT) bootstrap() {
	// Todo:
	this.syncAddNodes(this.seeds)
	this.DisplayRoutingTable()

	fmt.Println("start lookup")
	this.lookup(this.nodeID)
}

// add node to routing table in synchronize
func (this *DHT) syncAddNodes(nodes []*types.Node) {
	waitRequestIds := make([]types.RequestId, 0)
	for _, seed := range nodes {
		addr, err := getNodeUDPAddr(seed)
		if err != nil {
			log.Infof("seed node %s address is error!", seed.ID)
			continue
		}
		requestId, isNewRequest := this.messagePool.AddRequest(seed,
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
	log.Info("refreshRoutingTable")
	// Todo:
	this.syncAddNodes(this.seeds)
	this.lookup(this.nodeID)

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
				// ping this node
				addr, err := getNodeUDPAddr(n)
				if err != nil {
					continue

				}
				// ping and wait node one by one
				reqId, isNewRequest := this.messagePool.AddRequest(n, types.DHT_PING_REQUEST, nil, true)
				if isNewRequest {
					this.ping(addr)
				}
				this.messagePool.Wait([]types.RequestId{reqId})
				closestNodes = addClosestNode(closestNodes, n, targetID)
			}
		}
	}

}

// addClosestNode adds a node to the closest list
func addClosestNode(closestNodes []*types.Node, n *types.Node, targetID types.NodeID) []*types.Node {
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
	return closestNodes
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
	findNodePayload := mt.FindNodePayload{
		FromID:   this.nodeID,
		TargetID: targetID,
	}
	findNodePacket, err := msgpack.NewFindNode(findNodePayload)
	if err != nil {
		log.Error("failed to new dht find node packet", err)
		return err
	}
	this.send(addr, findNodePacket)
	return nil
}

// addNode adds a node to the K bucket.
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
}

// ping the remote node
func (this *DHT) ping(addr *net.UDPAddr) error {
	pingPayload := mt.DHTPingPayload{
		Version: this.version,
	}
	copy(pingPayload.FromID[:], this.nodeID[:])

	pingPayload.SrcEndPoint.UDPPort = this.udpPort
	pingPayload.SrcEndPoint.TCPPort = this.tcpPort

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(pingPayload.SrcEndPoint.Addr[:], ip[:16])

	pingPayload.DestEndPoint.UDPPort = uint16(addr.Port)

	destIP := addr.IP.To16()
	copy(pingPayload.DestEndPoint.Addr[:], destIP[:16])

	pingPacket, err := msgpack.NewDHTPing(pingPayload)
	if err != nil {
		log.Error("failed to new dht ping packet", err)
		return err
	}
	this.send(addr, pingPacket)
	return nil
}

// pong reply remote node when receiving ping
func (this *DHT) pong(addr *net.UDPAddr) error {
	PongPayload := mt.DHTPongPayload{
		Version: this.version,
	}
	copy(PongPayload.FromID[:], this.nodeID[:])
	PongPayload.SrcEndPoint.UDPPort = this.udpPort
	PongPayload.SrcEndPoint.TCPPort = this.tcpPort

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(PongPayload.SrcEndPoint.Addr[:], ip[:16])

	PongPayload.DestEndPoint.UDPPort = uint16(addr.Port)
	destIP := addr.IP.To16()
	copy(PongPayload.DestEndPoint.Addr[:], destIP[:16])

	pongPacket, err := msgpack.NewDHTPong(PongPayload)
	if err != nil {
		log.Error("failed to new dht pong packet", err)
		return err
	}
	this.send(addr, pongPacket)
	return nil
}

// findNodeReply reply remote node when receiving find node
func (this *DHT) findNodeReply(addr *net.UDPAddr, targetId types.NodeID) error {
	// query routing table
	nodes := this.routingTable.getClosestNodes(types.BUCKET_SIZE, targetId)
	neighborsPayload := mt.NeighborsPayload{
		FromID: this.nodeID,
		Nodes:  make([]types.Node, 0, len(nodes)),
	}
	for _, node := range nodes {
		log.Infof("ReturnNeightbors: %s:%d", node.IP, node.UDPPort)
		neighborsPayload.Nodes = append(neighborsPayload.Nodes, *node)
	}
	neighborsPacket, err := msgpack.NewNeighbors(neighborsPayload)
	if err != nil {
		log.Error("failed to new dht neighbors packet", err)
		return err
	}
	this.send(addr, neighborsPacket)
	return nil
}

// processPacket invokes the related handler to process the packet
func (this *DHT) processPacket(from *net.UDPAddr, packet []byte) {
	msgType, err := mt.MsgType(packet)
	if err != nil {
		log.Info("failed to get msg type")
		return
	}

	log.Infof("Recv UDP msg %s %v", msgType, from)
	switch msgType {
	case "DHTPing":
		this.pingHandle(from, packet)
	case "DHTPong":
		this.pongHandle(from, packet)
	case "findnode":
		this.findNodeHandle(from, packet)
	case "neighbors":
		this.neighborsHandle(from, packet)
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
			fmt.Printf("%x %s %d %d\n", bucket.entries[i].ID[:10], bucket.entries[i].IP,
				bucket.entries[i].UDPPort, bucket.entries[i].TCPPort)
		}
	}
}
