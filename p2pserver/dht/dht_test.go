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
	"crypto/rand"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Init(log.PATH, log.Stdout)
}

func createNodes(num int) []*types.Node {
	nodes := make([]*types.Node, 0, num)
	for i := 0; i < num; i++ {
		//node := config.DefConfig.Genesis.DHT.Seeds[i]
		node := &types.Node{
			IP:      "127.0.0.1",
			UDPPort: uint16(30000 + i),
			TCPPort: uint16(20000 + i),
		}
		id := types.ConstructID(node.IP, node.UDPPort)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, id)
		copy(node.ID[:], b[:])
		nodes = append(nodes, node)
	}
	return nodes
}

func TestDHT(t *testing.T) {
	id := types.ConstructID("127.0.0.1", 20332)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID types.NodeID
	copy(nodeID[:], b[:])

	dht := NewDHT(nodeID)
	assert.NotNil(t, dht)
	assert.NotNil(t, dht.routingTable)
	assert.NotNil(t, dht.messagePool)
	assert.NotNil(t, dht.recvCh)
	assert.NotNil(t, dht.bootstrapNodes)
	assert.NotNil(t, dht.feedCh)
	assert.Equal(t, nodeID, dht.nodeID)

	tempID := binary.LittleEndian.Uint64(dht.nodeID[:])
	assert.Equal(t, id, tempID)

	dht.addr = "127.0.0.1"
	dht.udpPort = 20332

	go dht.Start()

	time.Sleep(1 * time.Second)
	assert.Empty(t, dht.bootstrapNodes)
	assert.NotNil(t, dht.conn)
	assert.Empty(t, dht.routingTable.totalNodes())

	nodes := createNodes(10)
	assert.Equal(t, 10, len(nodes))

	for _, node := range nodes {
		dht.addNode(node)
	}
	assert.NotEmpty(t, dht.routingTable.totalNodes())

	tempID = binary.LittleEndian.Uint64(nodes[0].ID[:])
	ret := dht.Resolve(tempID)
	assert.Equal(t, 1, len(ret))

	tempID = 13494701530128158541
	ret = dht.Resolve(tempID)
	assert.Equal(t, types.FACTOR, len(ret))

	for _, node := range nodes {
		dht.routingTable.removeNode(node.ID)
	}
	assert.Empty(t, dht.routingTable.totalNodes())

	for _, node := range nodes {
		srcAddr := new(net.UDPAddr)
		srcAddr.IP = net.ParseIP(node.IP).To16()
		assert.NotNil(t, srcAddr.IP)
		srcAddr.Port = int(node.UDPPort)

		targetAddr := new(net.UDPAddr)
		targetAddr.IP = net.ParseIP(dht.addr).To16()
		assert.NotNil(t, targetAddr.IP)
		targetAddr.Port = int(dht.udpPort)

		// Ping
		pingMsg := msgpack.NewDHTPing(node.ID, node.UDPPort, node.TCPPort, srcAddr.IP, targetAddr, 1)
		pingBuffer := new(bytes.Buffer)
		mt.WriteMessage(pingBuffer, pingMsg)

		ping := &types.DHTMessage{
			From:    srcAddr,
			Payload: pingBuffer.Bytes(),
		}
		dht.recvCh <- ping

		// Findnode
		var targetID types.NodeID
		rand.Read(targetID[:])
		findNodeMsg := msgpack.NewFindNode(node.ID, targetID)
		findNodeBuffer := new(bytes.Buffer)
		mt.WriteMessage(findNodeBuffer, findNodeMsg)
		findNode := &types.DHTMessage{
			From:    srcAddr,
			Payload: findNodeBuffer.Bytes(),
		}
		dht.recvCh <- findNode

		// Pong
		pongMsg := msgpack.NewDHTPong(node.ID, node.UDPPort, node.TCPPort, srcAddr.IP, targetAddr, 1)
		pongBuffer := new(bytes.Buffer)
		mt.WriteMessage(pongBuffer, pongMsg)
		pong := &types.DHTMessage{
			From:    srcAddr,
			Payload: pongBuffer.Bytes(),
		}
		dht.recvCh <- pong
	}

	time.Sleep(3 * time.Second)
	assert.Empty(t, dht.recvCh)

	dht.Stop()
}
