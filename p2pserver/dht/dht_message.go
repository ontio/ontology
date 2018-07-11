package dht

import (
	"bytes"
	"errors"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"net"
	"strconv"
	"sync"
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

	requestId := types.ConstructRequestId(findNode.FromID, types.DHT_FIND_NODE_REQUEST)
	log.Info("receive find_node of", requestId)
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

	requestId := types.ConstructRequestId(neighbors.FromID, types.DHT_FIND_NODE_REQUEST)
	this.messagePool.DeleteRequest(requestId)

	pingReqIds := make([]types.RequestId, 0)

	waitGroup := new(sync.WaitGroup)
	for i := 0; i < len(neighbors.Nodes); i++ {
		node := &neighbors.Nodes[i]
		nodeAddress := node.IP + ":" + strconv.Itoa(int(node.UDPPort))
		if this.isInBlackList(nodeAddress) {
			continue
		}
		if node.ID == this.nodeID {
			continue
		}
		// ping this node
		addr, err := getNodeUDPAddr(node)
		if err != nil {
			continue
		}
		reqId, isNewRequest := this.messagePool.AddRequest(node, types.DHT_PING_REQUEST, nil, waitGroup)
		if isNewRequest {
			this.ping(addr)
		}
		pingReqIds = append(pingReqIds, reqId)
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
	log.Info("receive neighbors of", requestId)
}

// pingHandle handles a ping message from UDP network
func (this *DHT) pingHandle(from *net.UDPAddr, msg mt.Message) {
	// black list detect
	nodeAddress := string(from.IP) + ":" + strconv.Itoa(from.Port)
	if this.isInBlackList(nodeAddress) {
		return
	}
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

	// add the node to routing table
	var node *types.Node
	if node = this.routingTable.queryNode(ping.FromID); node == nil {
		node = &types.Node{
			ID:      ping.FromID,
			IP:      from.IP.String(),
			UDPPort: uint16(from.Port),
			TCPPort: uint16(ping.SrcEndPoint.TCPPort),
		}
	}
	this.addNode(node)
	this.pong(from)
	this.DisplayRoutingTable()
	requestId := types.ConstructRequestId(ping.FromID, types.DHT_PING_REQUEST)
	log.Info("receive ping of", requestId)
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

	requestId := types.ConstructRequestId(pong.FromID, types.DHT_PING_REQUEST)
	node, ok := this.messagePool.GetRequestData(requestId)
	if !ok {
		// request pool doesn't contain the node, ping timeout
		this.routingTable.removeNode(pong.FromID)
		return
	}

	// add to routing table
	this.addNode(node)
	// remove node from request pool
	this.messagePool.DeleteRequest(requestId)
	log.Info("receive pong of", requestId)
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
	return nil
}

// findNodeReply reply remote node when receiving find node
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

	return nil
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
	mt.WriteMessage(bf, pingMsg)
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

// onRequestTimeOut handles a timeout event of request
func (this *DHT) onRequestTimeOut(requestId types.RequestId) {
	reqType := types.GetReqTypeFromReqId(requestId)
	this.messagePool.DeleteRequest(requestId)
	log.Info("request", requestId, "timeout!")
	if reqType == types.DHT_FIND_NODE_REQUEST {
		results := make([]*types.Node, 0)
		this.messagePool.SetResults(results)
	} else if reqType == types.DHT_PING_REQUEST {
		replaceNode, ok := this.messagePool.GetReplaceNode(requestId)
		if ok && replaceNode != nil {
			bucketIndex, _ := this.routingTable.locateBucket(replaceNode.ID)
			this.routingTable.addNode(replaceNode, bucketIndex)
		}
	}
}
