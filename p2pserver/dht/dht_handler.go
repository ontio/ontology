package dht

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

// findNodeHandle handles a find node message from UDP network
func (this *DHT) findNodeHandle(from *net.UDPAddr, msg mt.Message) {
	findNode, ok := msg.(*mt.FindNode)
	if !ok {
		log.Error("find node handle detected error message type!")
		return
	}

	if node := this.routingTable.queryNode(findNode.FromID); node == nil {
		return
	}

	this.updateNode(findNode.FromID)
	this.findNodeReply(from, findNode.TargetID)
}

// neighborsHandle handles a neighbors message from UDP network
func (this *DHT) neighborsHandle(from *net.UDPAddr, msg mt.Message) {
	neighbors, ok := msg.(*mt.Neighbors)
	if !ok {
		log.Error("neighbors handle detected error message type!")
		return
	}
	if node := this.routingTable.queryNode(neighbors.FromID); node == nil {
		return
	}

	requestId := types.ConstructRequestId(neighbors.FromID,
		types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	pingReqIds := make([]types.RequestId, 0)

	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
		if node.ID == this.nodeID {
			continue
		}
		// ping this node
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			continue
		}
		reqId, isNewRequest := this.messagePool.AddRequest(node, types.DHT_PING_REQUEST, nil, true)
		if isNewRequest {
			this.ping(addr)
		}
		pingReqIds = append(pingReqIds, reqId)
	}
	this.messagePool.Wait(pingReqIds)
	liveNodes := make([]*types.Node, 0)
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
		if queryResult := this.routingTable.queryNode(node.ID); queryResult != nil {
			liveNodes = append(liveNodes, node)
		}
	}
	this.messagePool.SetResults(liveNodes)

	this.updateNode(neighbors.FromID)
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, msg mt.Message) {
	ping, ok := msg.(*mt.DHTPing)
	if !ok {
		log.Error("ping handle detected error message type!")
		return
	}
	if ping.Version != this.version {
		log.Errorf("pingHandle: version is incompatible. local %d remote %d",
			this.version, ping.Version)
		return
	}

	// if routing table doesn't contain the node, add it to routing table and wait request return
	if node := this.routingTable.queryNode(ping.FromID); node == nil {
		node := &types.Node{
			ID:      ping.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(from.Port),
			TCPPort: uint16(ping.SrcEndPoint.TCPPort),
		}
		this.addNode(node)
	} else {
		// update this node
		bucketIndex, _ := this.routingTable.locateBucket(ping.FromID)
		this.routingTable.addNode(node, bucketIndex)
	}
	this.pong(from)
	this.DisplayRoutingTable()
}

// pongHandle handles a pong message from UDP network
func (this *DHT) pongHandle(from *net.UDPAddr, msg mt.Message) {
	pong, ok := msg.(*mt.DHTPong)
	if !ok {
		log.Error("pong handle detected error message type!")
		return
	}
	if pong.Version != this.version {
		log.Errorf("pongHandle: version is incompatible. local %d remote %d",
			this.version, pong.Version)
		return
	}

	requesetId := types.ConstructRequestId(pong.FromID, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requesetId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.removeNode(pong.FromID)
		return
	}

	// add to routing table
	this.addNode(node)
	// remove node from request pool
	this.messagePool.DeleteRequest(requesetId)
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
