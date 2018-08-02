package types

import (
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func genTestFindNode() *FindNode {
	seed := genTestNode()
	findNode := new(FindNode)
	findNode.FromID = seed.ID
	findNode.TargetID = seed.ID
	return findNode
}

func TestFindNode(t *testing.T) {
	findNode := genTestFindNode()
	assert.Equal(t, common.DHT_FIND_NODE, findNode.CmdType())
	bf, err := findNode.Serialization()
	assert.Nil(t, err)

	deseFindNode := new(FindNode)
	err = deseFindNode.Deserialization(bf)
	assert.Nil(t, err)
	assert.Equal(t, findNode, deseFindNode)

	MessageTest(t, findNode)
}
