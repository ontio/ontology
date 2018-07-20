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
	"encoding/binary"
	"testing"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Init(log.PATH, log.Stdout)
}

func TestRoutingTable(t *testing.T) {
	id := types.ConstructID("127.0.0.1", 20332)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID types.NodeID
	copy(nodeID[:], b[:])

	routingTable := &routingTable{}
	routingTable.init(nodeID, nil)

	assert.Equal(t, nodeID, routingTable.id)
	assert.Nil(t, routingTable.feedCh)
	assert.NotEmpty(t, routingTable.buckets)

	for _, bucket := range routingTable.buckets {
		assert.Equal(t, 0, len(bucket.entries))
	}

	nodes := createNodes(100)
	assert.Equal(t, 100, len(nodes))

	// Add nodes to buckets
	for _, node := range nodes {
		bucketIndex, _ := routingTable.locateBucket(node.ID)
		remoteNode, _ := routingTable.isNodeInBucket(node.ID, bucketIndex)
		assert.Nil(t, remoteNode)

		num := routingTable.getTotalNodeNumInBukcet(bucketIndex)

		added := routingTable.addNode(node, bucketIndex)
		if num >= types.BUCKET_SIZE {
			assert.Equal(t, false, added)
			_, ok := routingTable.isNodeInBucket(node.ID, bucketIndex)
			assert.Equal(t, false, ok)
		} else {
			assert.Equal(t, true, added)
			_, ok := routingTable.isNodeInBucket(node.ID, bucketIndex)
			assert.Equal(t, true, ok)

			frontNode := routingTable.buckets[bucketIndex].entries[0]
			assert.NotNil(t, frontNode)
			assert.Equal(t, frontNode.ID, node.ID)
			assert.Equal(t, frontNode.IP, node.IP)
			assert.Equal(t, frontNode.UDPPort, node.UDPPort)
			assert.Equal(t, frontNode.TCPPort, node.TCPPort)
		}
	}

	// Remove nodes from buckets
	for _, node := range nodes {
		routingTable.removeNode(node.ID)
	}

	for _, bucket := range routingTable.buckets {
		assert.Equal(t, 0, len(bucket.entries))
	}
}
