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

package test

import (
	"encoding/binary"
	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"testing"
)

func TestCon(t *testing.T) {
	id := types.ConstructID("127.0.0.1", uint16(18888))
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID types.NodeID
	copy(nodeID[:], b[:])
	reqId := types.ConstructRequestId(nodeID, types.DHT_FIND_NODE_REQUEST)
	reqType := types.GetReqTypeFromReqId(reqId)
	assert.Equal(t, reqType, types.DHT_FIND_NODE_REQUEST)
}
