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
