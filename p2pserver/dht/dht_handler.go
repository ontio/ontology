package dht

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
)

func (this *DHT) FindNodeHandler(from *net.UDPAddr, findNodeMsgData []byte) error {
	findNodeMsg := new(mt.FindNode)
	findNodeMsg.Deserialization(findNodeMsgData)
	// query routing table
	nodes := this.routingTable.GetClosestNodes(types.BUCKET_SIZE, findNodeMsg.P.TargetID)
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
	this.send(from, neighborsPacket)
	return nil
}

func (this *DHT) NeighborsHandler(from *net.UDPAddr, neighborsMsgData []byte) {
	neighborsMsg := new(mt.Neighbors)
	neighborsMsg.Deserialization(neighborsMsgData)
	neighbors := neighborsMsg.P.Nodes
	for _, neighbor := range neighbors {
		if this.routingTable.queryNode(neighbor.ID) == nil {
			// add neighbors to routing table
			this.routingTable.AddNode(&neighbor)
			// send find node request to this neighbor
			_, err := getNodeUdpAddr(&neighbor)
			if err != nil {
				continue
			}
			//this.FindNode(addr, 0)
		}
	}
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