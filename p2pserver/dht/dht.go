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

type DHT struct {
	version      uint16
	nodeID       types.NodeID
	mu           sync.Mutex
	routingTable *routingTable
	addr         string
	udpPort      uint16
	tcpPort      uint16
	conn         *net.UDPConn
	messagePool  *types.DHTMessagePool
	recvCh       chan *types.DHTMessage
	stopCh       chan struct{}
	seeds        []*types.Node
	feedCh       chan *types.FeedEvent
}

func NewDHT(node types.NodeID, seeds []*types.Node) *DHT {
	// Todo:
	dht := &DHT{
		nodeID:       node,
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

func (this *DHT) init() {
	this.recvCh = make(chan *types.DHTMessage, types.MSG_CACHE)
	this.stopCh = make(chan struct{})
	this.messagePool = types.NewRequestPool(this.onRequestTimeOut)
	this.feedCh = make(chan *types.FeedEvent, types.MSG_CACHE)
	this.routingTable.init(this.nodeID, this.feedCh)
}

func (this *DHT) Start() {
	go this.Loop()

	err := this.ListenUDP(this.addr + ":" + strconv.Itoa(int(this.udpPort)))
	if err != nil {
		log.Errorf("listen udp failed.")
	}
	this.Bootstrap()
}

func (this *DHT) Stop() {
	if this.stopCh != nil {
		this.stopCh <- struct{}{}
	}

	if this.feedCh != nil {
		close(this.feedCh)
	}
}

func (this *DHT) Bootstrap() {
	// Todo:
	this.SyncAddNodes(this.seeds)
	this.DisplayRoutingTable()

	fmt.Println("start lookup")
	this.lookup(this.nodeID)
}

// add node to routing table in synchronize
func (this *DHT) SyncAddNodes(nodes []*types.Node) {
	//log.Infof("SyncAddNodes start")
	waitRequestIds := make([]types.RequestId, 0)
	for _, seed := range nodes {
		addr, err := getNodeUDPAddr(seed)
		if err != nil {
			fmt.Printf("seed node %s address is error!", seed.ID)
			continue
		}
		requestId, isNewRequest := this.messagePool.AddRequest(seed, types.DHT_PING_REQUEST, nil, true)
		if isNewRequest {
			this.Ping(addr)
		}
		waitRequestIds = append(waitRequestIds, requestId)
	}
	//log.Infof("SyncAddNodes completed")
	this.messagePool.Wait(waitRequestIds)
}

func (this *DHT) GetFeedCh() chan *types.FeedEvent {
	return this.feedCh
}

func (this *DHT) Loop() {
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

func (this *DHT) refreshRoutingTable() {
	log.Info("refreshRoutingTable")
	// Todo:
	this.SyncAddNodes(this.seeds)
	this.lookup(this.nodeID)

	var targetID types.NodeID
	rand.Read(targetID[:])
	log.Infof("refreshRoutingTable: %s", targetID.String())
	this.lookup(targetID)
}

func (this *DHT) lookup(targetID types.NodeID) []*types.Node {
	bucket, _ := this.routingTable.locateBucket(targetID)
	node, ret := this.routingTable.isNodeInBucket(targetID, bucket)
	if ret == true {
		log.Infof("targetID %s is in the bucket %d", targetID.String(), bucket)
		return []*types.Node{node}
	}

	closestNodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, targetID)
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
				this.FindNode(node, targetID)
				this.messagePool.AddRequest(node, types.DHT_FIND_NODE_REQUEST, nil, false)
			}()
		}

		if pendingQueries == 0 {
			break
		}

		//log.Info("Waiting for response")

		this.waitAndHandleResponse(knownNode, closestNodes, targetID)

		pendingQueries--
	}
	return closestNodes
}

func (this *DHT) waitAndHandleResponse(knownNode map[types.NodeID]bool, closestNodes []*types.Node, targetID types.NodeID) {
	responseCh := this.messagePool.GetResultChan()
	select {
	case entries, ok := <-responseCh:
		if ok {
			//log.Infof("get entries %d %v", len(entries), entries)
			for _, n := range entries {
				log.Infof("receive new node %d, id %s", n.UDPPort, n.ID.String())
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
				if isNewRequest{
					this.Ping(addr)
				}
				this.messagePool.Wait([]types.RequestId{reqId})
				closestNodes = addClosestNode(closestNodes, n, targetID)
			}
		}
	}

}

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

func (this *DHT) onRequestTimeOut(requestId types.RequestId) {
	reqType := types.GetReqTypeFromReqId(requestId)
	this.messagePool.DeleteRequest(requestId)
	fmt.Println("request ", requestId, "timeout!")
	if reqType == types.DHT_FIND_NODE_REQUEST {
		results := make([]*types.Node, 0)
		this.messagePool.SetResults(results)
	} else if reqType == types.DHT_PING_REQUEST {
		pendingNode, ok := this.messagePool.GetSupportData(requestId)
		if ok && pendingNode != nil {
			bucketIndex, _ := this.routingTable.locateBucket(pendingNode.ID)
			this.routingTable.AddNode(pendingNode, bucketIndex)
		}
	}
}

func (this *DHT) FindNode(remotePeer *types.Node, targetID types.NodeID) error {
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

func (this *DHT) AddNode(remotePeer *types.Node) {
	if remotePeer == nil || remotePeer.ID == this.nodeID {
		return
	}

	// find node in own bucket
	bucketIndex, _ := this.routingTable.locateBucket(remotePeer.ID)
	remoteNode, isInBucket := this.routingTable.isNodeInBucket(remotePeer.ID, bucketIndex)
	// update peer info in local bucket
	remoteNode = remotePeer
	if isInBucket {
		this.routingTable.AddNode(remoteNode, bucketIndex)
	} else {
		bucketNodeNum := this.routingTable.GetTotalNodeNumInBukcet(bucketIndex)
		if bucketNodeNum < types.BUCKET_SIZE { // bucket is not full
			this.routingTable.AddNode(remoteNode, bucketIndex)
		} else {
			lastNode := this.routingTable.GetLastNodeInBucket(bucketIndex)
			addr, err := getNodeUDPAddr(lastNode)
			if err != nil {
				this.routingTable.RemoveNode(lastNode.ID)
				this.routingTable.AddNode(remoteNode, bucketIndex)
				return
			}
			if _, isNewRequest := this.messagePool.AddRequest(lastNode, types.DHT_PING_REQUEST, remotePeer,
				false); isNewRequest {
				this.Ping(addr)
			}
		}
	}
}

func (this *DHT) Ping(addr *net.UDPAddr) error {
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

func (this *DHT) Pong(addr *net.UDPAddr) error {
	PongPayload := mt.DHTPongPayload{
		Version: this.version,
	}
	copy(PongPayload.FromID[:], this.nodeID[:])
	log.Infof("Pong: fromID %s", this.nodeID.String())
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

// response to find node
func (this *DHT) FindNodeReply(addr *net.UDPAddr, targetId types.NodeID) error {
	// query routing table
	nodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, targetId)
	neighborsPayload := mt.NeighborsPayload{
		FromID: this.nodeID,
		Nodes:  make([]types.Node, 0, len(nodes)),
	}
	log.Infof("ReturenNeighbors: nodes %d", len(nodes))
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

func (this *DHT) processPacket(from *net.UDPAddr, packet []byte) {
	// Todo: add processPacket implementation
	msgType, err := mt.MsgType(packet)
	if err != nil {
		log.Info("failed to get msg type")
		return
	}

	log.Infof("Recv UDP msg %s %v", msgType, from)
	switch msgType {
	case "DHTPing":
		this.PingHandler(from, packet)
	case "DHTPong":
		this.PongHandler(from, packet)
	case "findnode":
		this.FindNodeHandler(from, packet)
	case "neighbors":
		this.NeighborsHandler(from, packet)
	default:
		log.Infof("processPacket: unknown msg %s", msgType)
	}
}

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

func (this *DHT) ListenUDP(laddr string) error {
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
	fmt.Println("DHT is listening on ", laddr)
	go this.recvUDPMsg()
	return nil
}

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
		if this.routingTable.GetTotalNodeNumInBukcet(bucketIndex) == 0 {
			continue
		}
		fmt.Println("[", bucketIndex, "]: ")
		for i := 0; i < this.routingTable.GetTotalNodeNumInBukcet(bucketIndex); i++ {
			fmt.Printf("%x \n", bucket.entries[i].ID[20:])
		}
	}
}
