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
	"bytes"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
)

// findNodeHandle handles a find node message from UDP network
func (this *DHT) findNodeHandle(from *net.UDPAddr, msg mt.Message) {
	findNode, ok := msg.(*mt.FindNode)
	if !ok {
		log.Error("[dht]find node handle detected error message type!")
		return
	}

	if node := this.routingTable.queryNode(findNode.FromID); node == nil {
		// findnode must be after ping/pong, in case of DoS attack
		log.Infof("[dht]findNodeHandle: from %v, local doesn't content the request node!", from)
		return
	}

	this.updateNode(findNode.FromID)
	this.findNodeReply(from, findNode.TargetID)
	log.Infof("[dht]findNodeHandle: from %v ", from)
}

// neighborsHandle handles a neighbors message from UDP network
func (this *DHT) neighborsHandle(from *net.UDPAddr, msg mt.Message) {
	neighbors, ok := msg.(*mt.Neighbors)
	if !ok {
		log.Error("[dht]neighbors handle detected error message type!")
		return
	}
	if node := this.routingTable.queryNode(neighbors.FromID); node == nil {
		return
	}

	requestId := types.ConstructRequestId(neighbors.FromID, types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	waitGroup := new(sync.WaitGroup)
	for i := 0; i < len(neighbors.Nodes) && i < types.BUCKET_SIZE; i++ {
		node := &neighbors.Nodes[i]
		if this.isInBlackList(node.IP) || !this.isInWhiteList(node.IP) {
			continue
		}
		if node.ID == this.nodeID {
			continue
		}

		whiteListLen := len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers)
		if config.DefConfig.P2PNode.ReservedPeersOnly && whiteListLen > 0 {
			var index = 0
			for ; index < whiteListLen; index++ {
				ip := config.DefConfig.P2PNode.ReservedCfg.ReservedPeers[index]
				if strings.HasPrefix(node.IP, ip) {
					break
				}
			}
			if index == whiteListLen {
				continue
			}
		}
		// ping this node
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			continue
		}
		_, isNewRequest := this.messagePool.AddRequest(node, types.DHT_PING_REQUEST, nil, waitGroup)
		if isNewRequest {
			this.ping(addr)
		}
	}
	waitGroup.Wait()
	liveNodes := make([]*types.Node, 0)
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
		if queryResult := this.routingTable.queryNode(node.ID); queryResult != nil {
			liveNodes = append(liveNodes, node)
		}
	}
	this.messagePool.SetResults(liveNodes)

	this.updateNode(neighbors.FromID)
	log.Infof("[dht]neighborsHandle: from %v ", from)
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, msg mt.Message) {
	ping, ok := msg.(*mt.DHTPing)
	if !ok {
		log.Error("[dht]ping handle detected error message type!")
		return
	}
	if ping.Version != this.version {
		log.Errorf("[dht]pingHandle: version is incompatible. local %d remote %d",
			this.version, ping.Version)
		return
	}

	if ping.DestEndPoint.UDPPort != this.udpPort {
		log.Errorf("[dht]pingHandle: udp port mismatch. input %d expected %d",
			ping.DestEndPoint.UDPPort, this.udpPort)
		return
	}

	// add the node to routing table
	var node *types.Node
	if node = this.routingTable.queryNode(ping.FromID); node == nil {
		node = &types.Node{
			ID:      ping.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(ping.SrcEndPoint.UDPPort),
			TCPPort: uint16(ping.SrcEndPoint.TCPPort),
		}
	}
	this.addNode(node)
	this.pong(from)
	log.Infof("[dht]pingHandle: from %v ", from)
}

// pongHandle handles a pong message from UDP network
func (this *DHT) pongHandle(from *net.UDPAddr, msg mt.Message) {
	pong, ok := msg.(*mt.DHTPong)
	if !ok {
		log.Error("[dht]pong handle detected error message type!")
		return
	}
	if pong.Version != this.version {
		log.Errorf("[dht]pongHandle: version is incompatible. local %d remote %d",
			this.version, pong.Version)
		return
	}

	requestId := types.ConstructRequestId(pong.FromID, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requestId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		log.Infof("[dht]pongHandle: from %v timeout", from)
		this.routingTable.removeNode(pong.FromID)
		return
	}

	if pong.SrcEndPoint.UDPPort != node.UDPPort ||
		pong.SrcEndPoint.TCPPort != node.TCPPort {
		log.Errorf("[dht]pongHandle: mismatch port: input: udp %d tcp %d, expected: udp %d tcp %d",
			pong.SrcEndPoint.UDPPort, pong.SrcEndPoint.TCPPort,
			node.UDPPort, node.TCPPort)
		return
	}

	// add to routing table
	this.addNode(node)
	// remove node from request pool
	this.messagePool.DeleteRequest(requestId)
	log.Infof("[dht]pongHandle: from %v ", from)
}

// update the node to bucket when receive message from the node
func (this *DHT) updateNode(fromId types.NodeID) {
	node := this.routingTable.queryNode(fromId)
	if node != nil {
		// add node to bucket
		bucketIndex, _ := this.routingTable.locateBucket(fromId)
		this.routingTable.addNode(node, bucketIndex)
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
	log.Infof("[dht]findNode to %s", addr.String())
	return nil
}

// findNodeReply replies remote node when receiving find node
func (this *DHT) findNodeReply(addr *net.UDPAddr, targetId types.NodeID) error {
	// query routing table
	nodes := this.routingTable.getClosestNodes(types.BUCKET_SIZE, targetId)

	maskPeers := config.DefConfig.P2PNode.ReservedCfg.MaskPeers
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(maskPeers) > 0 {
		for i := 0; i < len(nodes); i++ {
			for j := 0; j < len(maskPeers); j++ {
				if nodes[i].IP == maskPeers[j] {
					nodes = append(nodes[:i], nodes[i+1:]...)
					i--
					break
				}
			}
		}
	}

	neighborsMsg := msgpack.NewNeighbors(this.nodeID, nodes)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, neighborsMsg)
	this.send(addr, bf.Bytes())
	log.Infof("[dht]findNodeReply to %s", addr.String())

	return nil
}

// ping the remote node
func (this *DHT) ping(addr *net.UDPAddr) error {
	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("[dht]Parse IP address error\n", this.addr)
		return errors.New("[dht]Parse IP address error")
	}
	pingMsg := msgpack.NewDHTPing(this.nodeID, this.udpPort,
		this.tcpPort, ip, addr, this.version)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, pingMsg)
	this.send(addr, bf.Bytes())
	log.Infof("[dht]ping to %s", addr.String())
	return nil
}

// pong reply remote node when receiving ping
func (this *DHT) pong(addr *net.UDPAddr) error {

	ip := net.ParseIP(this.addr).To16()
	if ip == nil {
		log.Error("[dht]Parse IP address error\n", this.addr)
		return errors.New("[dht]Parse IP address error")
	}

	pongMsg := msgpack.NewDHTPong(this.nodeID, this.udpPort,
		this.tcpPort, ip, addr, this.version)
	bf := new(bytes.Buffer)
	mt.WriteMessage(bf, pongMsg)
	this.send(addr, bf.Bytes())
	log.Infof("[dht]pong to %s", addr.String())
	return nil
}

// onRequestTimeOut handles a timeout event of request
func (this *DHT) onRequestTimeOut(requestId types.RequestId) {
	reqType := types.GetReqTypeFromReqId(requestId)
	if reqType == types.DHT_FIND_NODE_REQUEST {
		results := make([]*types.Node, 0)
		this.messagePool.SetResults(results)
	} else if reqType == types.DHT_PING_REQUEST {
		replaceNode, ok := this.messagePool.GetReplaceNode(requestId)
		destNode, ok := this.messagePool.GetRequestData(requestId)
		if ok && destNode != nil{
			this.routingTable.removeNode(destNode.ID)
		}
		if ok && replaceNode != nil {
			bucketIndex, _ := this.routingTable.locateBucket(replaceNode.ID)
			this.routingTable.addNode(replaceNode, bucketIndex)
		}
	}
	this.messagePool.DeleteRequest(requestId)
}
