package dht

import (
	"fmt"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

func (this *DHT) FindNodeHandler(from *net.UDPAddr, findNodeMsgData []byte) error {
	findNodeMsg := new(mt.FindNode)
	findNodeMsg.Deserialization(findNodeMsgData)
	fromNode := this.routingTable.queryNode(findNodeMsg.P.FromID)
	if fromNode != nil {
		// add node to bucket
		this.AddNode(fromNode)
	}
	return this.FindNodeReply(from, findNodeMsg.P.TargetID)
}

func (this *DHT) NeighborsHandler(from *net.UDPAddr, neighborsMsgData []byte) {
	neighborsMsg := &mt.Neighbors{}
	neighborsMsg.Deserialization(neighborsMsgData)
	neighbors := neighborsMsg.P.Nodes

	requestId := types.ConstructRequestId(neighborsMsg.P.FromID, types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	fmt.Print("return neighbors to ", requestId, ", return result is: ")
	results := make([]*types.Node, 0, len(neighbors))
	for i := 0; i < len(neighbors); i++ {
		results = append(results, &neighbors[i])
		fmt.Print(neighbors[i].UDPPort, ",")
	}
	fmt.Println()
	this.messagePool.SetResults(results)

	fromNode := this.routingTable.queryNode(neighborsMsg.P.FromID)
	if fromNode != nil {
		// add node to bucket
		this.AddNode(fromNode)
	}
}

func (this *DHT) PingHandler(fromAddr *net.UDPAddr, pingMsgData []byte) {
	pingMsg := new(mt.DHTPing)
	pingMsg.Deserialization(pingMsgData)
	// response
	this.Pong(fromAddr)
	this.updateFromNode(pingMsg.P.FromID, fromAddr)
	fmt.Println("receive ping of ", fromAddr, "current DHT is: ")
	this.DisplayRoutingTable()
}

func (this *DHT) PongHandler(fromAddr *net.UDPAddr, pongMsgData []byte) {
	pongMsg := new(mt.DHTPong)
	pongMsg.Deserialization(pongMsgData)
	fromNodeId := pongMsg.P.FromID
	requesetId := types.ConstructRequestId(fromNodeId, types.DHT_PING_REQUEST)
	fromNode, ok := this.messagePool.GetRequestData(requesetId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.RemoveNode(fromNodeId)
		return
	}

	// add to routing table
	this.AddNode(fromNode)
	// remove node from request pool
	this.messagePool.DeleteRequest(requesetId)
	fmt.Println("receive pong of ", requesetId)
}

// update the node to bucket when receive message from the node
func (this *DHT) updateFromNode(fromNodeId types.NodeID, fromAddr *net.UDPAddr) {
	fromNode := this.routingTable.queryNode(fromNodeId)
	if fromNode != nil {
		// add node to bucket
		this.AddNode(fromNode)
	} else {
		node := &types.Node{
			ID:      fromNodeId,
			IP:      fromAddr.IP.String(),
			UDPPort: uint16(fromAddr.Port),
		}
		this.AddNode(node)
	}
}
