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
