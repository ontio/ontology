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

package types

import (
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func genTestNeighbors() *Neighbors {
	neighbors := new(Neighbors)
	node := genTestNode()
	neighbors.FromID = node.ID
	neighbors.Nodes = make([]types.Node, 0)
	neighbors.Nodes = append(neighbors.Nodes, *node)
	return neighbors
}

func TestNeighbors(t *testing.T) {
	neighbors := genTestNeighbors()
	assert.Equal(t, common.DHT_NEIGHBORS, neighbors.CmdType())
	bf, err := neighbors.Serialization()
	assert.Nil(t, err)

	deseNeighbors := new(Neighbors)
	err = deseNeighbors.Deserialization(bf)
	assert.Nil(t, err)
	assert.Equal(t, neighbors, deseNeighbors)

	MessageTest(t, neighbors)
}
