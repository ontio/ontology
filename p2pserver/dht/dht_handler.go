package dht

import (
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

func (this *DHT) FindNodeHandler(from *net.UDPAddr, findNodeMsgData []byte) error {
	findNodeMsg := new(mt.FindNode)
	findNodeMsg.Deserialization(findNodeMsgData)
	return this.ReturnNeighbors(from, findNodeMsg.P.TargetID)
}

func (this *DHT) NeighborsHandler(from *net.UDPAddr, neighborsMsgData []byte) {
	neighborsMsg := new(mt.Neighbors)
	neighborsMsg.Deserialization(neighborsMsgData)
	neighbors := neighborsMsg.P.Nodes
	results := make([]*types.Node, 0)
	for _, neighbor := range neighbors {
		results = append(results, &neighbor)
	}
	this.findNodeQueue.SetResult(results, neighborsMsg.P.FromID)
}

func (this *DHT) PingHandler(fromAddr *net.UDPAddr, pingMsgData []byte) {
	pingMsg := new(mt.DHTPing)
	pingMsg.Deserialization(pingMsgData)
	// response
	this.Pong(fromAddr)
	// add this node to bucket
	fromNodeId := pingMsg.P.FromID
	fromNode := this.routingTable.queryNode(fromNodeId)
	if fromNode != nil {
		// add node to bucket
		this.AddNode(fromNode)
	}
}

func (this *DHT) PongHandler(fromAddr *net.UDPAddr, pongMsgData []byte) {
	pongMsg := new(mt.DHTPong)
	pongMsg.Deserialization(pongMsgData)
	fromNodeId := pongMsg.P.FromID
	fromNode, ok := this.pingNodeQueue.GetRequestNode(fromNodeId)
	if !ok {
		// ping node queue doesn't contain the node, ping timeout
		this.routingTable.RemoveNode(fromNodeId)
	} else {
		// add to bucket header
		this.routingTable.AddNode(fromNode)
		// remove node from ping node queue
		this.pingNodeQueue.DeleteNode(fromNodeId)
	}
}
