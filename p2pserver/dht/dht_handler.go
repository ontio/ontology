package dht

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

// findNodeHandle handles a find node message from UDP network
func (this *DHT) findNodeHandle(from *net.UDPAddr, packet []byte) {
	var findNode mt.FindNode
	if err := findNode.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if node := this.routingTable.queryNode(findNode.P.FromID); node == nil {
		return
	}

	this.updateNode(findNode.P.FromID, from)
	this.findNodeReply(from, findNode.P.TargetID)
}

// neighborsHandle handles a neighbors message from UDP network
func (this *DHT) neighborsHandle(from *net.UDPAddr, packet []byte) {
	var neighbors mt.Neighbors
	if err := neighbors.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	if node := this.routingTable.queryNode(neighbors.P.FromID); node == nil {
		return
	}

	requestId := types.ConstructRequestId(neighbors.P.FromID,
		types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	results := make([]*types.Node, 0, len(neighbors.P.Nodes))
	for i := 0; i < len(neighbors.P.Nodes); i++ {
		results = append(results, &neighbors.P.Nodes[i])
	}
	this.messagePool.SetResults(results)

	this.updateNode(neighbors.P.FromID, from)
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, packet []byte) {
	var ping mt.DHTPing
	if err := ping.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}
	this.pong(from)

	if node := this.routingTable.queryNode(ping.P.FromID); node == nil {
		node := &types.Node{
			ID:      ping.P.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(from.Port),
			TCPPort: uint16(ping.P.SrcEndPoint.TCPPort),
		}
		this.addNode(node)
	} else {
		this.updateNode(ping.P.FromID, from)
	}

	this.DisplayRoutingTable()
}

// pongHandle handles a pong message from UDP network
func (this *DHT) pongHandle(from *net.UDPAddr, packet []byte) {
	var pong mt.DHTPong
	if err := pong.Deserialization(packet); err != nil {
		log.Error(err)
		return
	}

	fromId := pong.P.FromID
	requesetId := types.ConstructRequestId(fromId, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requesetId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.removeNode(fromId)
		return
	}

	// add to routing table
	this.addNode(node)
	// remove node from request pool
	this.messagePool.DeleteRequest(requesetId)
}

// update the node to bucket when receive message from the node
func (this *DHT) updateNode(fromId types.NodeID, from *net.UDPAddr) {
	node := this.routingTable.queryNode(fromId)
	if node != nil {
		// add node to bucket
		this.addNode(node)
	}
}
