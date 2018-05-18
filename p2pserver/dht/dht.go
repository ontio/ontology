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
	//"fmt"
	"errors"
	"net"
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

type DHT struct {
	version       uint16
	nodeID        types.NodeID
	mu            sync.Mutex
	routingTable  *routingTable
	addr          string
	port          uint16
	conn          *net.UDPConn
	pingNodeQueue *types.PingNodeQueue
	findNodeQueue *types.FindNodeQueue
	recvCh        chan *types.DHTMessage
	stopCh        chan struct{}
	seeds         []*types.Node
}

func NewDHT() *DHT {
	// Todo:
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID, _ := types.PubkeyID(pub)
	dht := &DHT{
		nodeID:       nodeID,
		routingTable: &routingTable{},
	}
	dht.init()
	return dht
}

func (this *DHT) init() {
	this.recvCh = make(chan *types.DHTMessage, types.MSG_CACHE)
	this.stopCh = make(chan struct{})
	this.pingNodeQueue = types.NewPingNodeQueue(this.onPingTimeOut)
	this.findNodeQueue = types.NewFindNodeQueue(this.onFindNodeTimeOut)
	this.routingTable.init(this.nodeID)
	this.seeds = make([]*types.Node, 0)

}

func (this *DHT) Start() {
	go this.Loop()

	err := this.ListenUDP("127.0.0.1:20334")
	if err != nil {
		log.Errorf("listen udp failed.")
	}
	this.Bootstrap()
}

func (this *DHT) Stop() {

}

func (this *DHT) Bootstrap() {
	// Todo:
	_, pub1, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID1, _ := types.PubkeyID(pub1)

	seed1 := &types.Node{
		ID:      nodeID1,
		IP:      "127.0.0.1",
		UDPPort: 20010,
		TCPPort: 20011,
	}
	this.seeds = append(this.seeds, seed1)
	this.AddNode(seed1)

	_, pub2, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID2, _ := types.PubkeyID(pub2)
	seed2 := &types.Node{
		ID:      nodeID2,
		IP:      "127.0.0.1",
		UDPPort: 30010,
		TCPPort: 30011,
	}
	this.seeds = append(this.seeds, seed2)
	this.AddNode(seed2)

	_, pub3, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID3, _ := types.PubkeyID(pub3)
	seed3 := &types.Node{
		ID:      nodeID3,
		IP:      "127.0.0.1",
		UDPPort: 40010,
		TCPPort: 40011,
	}
	this.seeds = append(this.seeds, seed3)
	this.AddNode(seed3)

	_, pub4, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID4, _ := types.PubkeyID(pub4)
	seed4 := &types.Node{
		ID:      nodeID4,
		IP:      "127.0.0.1",
		UDPPort: 50010,
		TCPPort: 50011,
	}
	this.seeds = append(this.seeds, seed4)
	this.AddNode(seed4)

	this.lookup(this.nodeID)
}

func (this *DHT) Loop() {
	for {
		select {
		case pk, ok := <-this.recvCh:
			if ok {
				this.processPacket(pk.From, pk.Payload)
			}
		case <-this.stopCh:
			return
		}
	}

}

func (this *DHT) lookup(targetID types.NodeID) []*types.Node {
	bucket, _ := this.routingTable.locateBucket(targetID)
	node, ret := this.routingTable.isNodeInBucket(targetID, bucket)
	if ret == true {
		log.Infof("targetID %s is in the bucket %d", targetID.String(), bucket)
		return []*types.Node{node}
	}

	visited := make(map[types.NodeID]bool)
	knownNode := make(map[types.NodeID]bool)
	pendingQueries := 0

	visited[this.nodeID] = true

	closestNodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, targetID)

	if len(closestNodes) == 0 {
		return nil
	}

	for _, node := range closestNodes {
		if node.ID == targetID {
			return closestNodes
		}
	}

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
				this.findNodeQueue.AddRequestNode(node)
				this.findNodeQueue.Timer(node.ID)
			}()
		}

		if pendingQueries == 0 {
			break
		}

		responseCh := this.findNodeQueue.GetResultCh()
		select {
		case entries, ok := <-responseCh:
			if ok {
				for _, n := range entries {
					log.Info("receive new node", n)
					// Todo:
					if knownNode[n.ID] == true {
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

		pendingQueries--
	}
	return closestNodes
}

func (this *DHT) FindNode(remotePeer *types.Node, targetID types.NodeID) error {
	addr, err := getNodeUdpAddr(remotePeer)
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

func (this *DHT) onFindNodeTimeOut(requestNodeId types.NodeID) {
	// remove the node from bucket
	this.routingTable.RemoveNode(requestNodeId)
	// push a empty slice to find node queue
	results := make([]*types.Node, 0)
	this.findNodeQueue.SetResult(results, requestNodeId)
}

func (this *DHT) AddNode(remotePeer *types.Node) {
	// find node in own bucket
	bucketIndex, _ := this.routingTable.locateBucket(remotePeer.ID)
	remoteNode, isInBucket := this.routingTable.isNodeInBucket(remotePeer.ID, bucketIndex)
	// update peer info in local bucket
	remoteNode = remotePeer
	if isInBucket {
		this.routingTable.AddNode(remoteNode)
	} else {
		bucketNodeNum := this.routingTable.GetTotalNodeNumInBukcet(bucketIndex)
		if bucketNodeNum < types.BUCKET_SIZE { // bucket is not full
			this.routingTable.AddNode(remoteNode)
		} else {
			lastNode := this.routingTable.GetLastNodeInBucket(bucketIndex)
			addr, err := getNodeUdpAddr(lastNode)
			if err != nil {
				this.routingTable.RemoveNode(lastNode.ID)
				this.routingTable.AddNode(remoteNode)
				return
			}
			this.pingNodeQueue.AddNode(lastNode, remoteNode, types.PING_TIMEOUT)
			this.Ping(addr)
		}
	}
}

func (this *DHT) Ping(addr *net.UDPAddr) error {
	pingPayload := mt.DHTPingPayload{
		Version:  this.version,
		SrcPort:  this.port,
		DestPort: uint16(addr.Port),
	}

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(pingPayload.SrcAddr[:], ip[:16])

	ip = addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	copy(pingPayload.DestAddr[:], ip[:16])

	copy(pingPayload.FromID[:], this.nodeID[:])

	pingPacket, err := msgpack.NewDHTPing(pingPayload)
	if err != nil {
		log.Error("failed to new dht ping packet", err)
		return err
	}
	this.send(addr, pingPacket)
	return nil
}

func (this *DHT) onPingTimeOut(nodeId types.NodeID) {
	// remove the node from bucket
	this.routingTable.RemoveNode(nodeId)
	pendingNode, ok := this.pingNodeQueue.GetPendingNode(nodeId)
	if ok && pendingNode != nil {
		// add pending node to bucket
		this.routingTable.AddNode(pendingNode)
	}
	// clear ping node queue
	this.pingNodeQueue.DeleteNode(nodeId)
}

func (this *DHT) Pong(addr *net.UDPAddr) error {
	PongPayload := mt.DHTPongPayload{
		Version:  this.version,
		SrcPort:  this.port,
		DestPort: uint16(addr.Port),
	}

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n", this.addr)
		return errors.New("Parse IP address error")
	}
	copy(PongPayload.SrcAddr[:], ip[:16])

	ip = addr.IP.To4()
	if ip == nil {
		ip = addr.IP.To16()
	}
	copy(PongPayload.DestAddr[:], ip[:16])

	copy(PongPayload.FromID[:], this.nodeID[:])

	pongPacket, err := msgpack.NewDHTPong(PongPayload)
	if err != nil {
		log.Error("failed to new dht pong packet", err)
		return err
	}
	this.send(addr, pongPacket)
	return nil
}

// response to find node
func (this *DHT) ReturnNeighbors(addr *net.UDPAddr, targetId types.NodeID) error {
	// query routing table
	nodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, targetId)
	neighborsPayload := mt.NeighborsPayload{
		FromID: this.nodeID,
	}
	for _, node := range nodes {
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

	log.Infof("Recv UDP msg %s", msgType)
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
			Payload: buf[:nbytes],
		}
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

func getNodeUdpAddr(node *types.Node) (*net.UDPAddr, error) {
	addr := new(net.UDPAddr)
	addr.IP = net.ParseIP(node.IP).To16()
	if addr.IP == nil {
		log.Error("Parse IP address error\n", node.IP)
		return nil, errors.New("Parse IP address error")
	}
	addr.Port = int(node.UDPPort)
	return addr, nil
}
