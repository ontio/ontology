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
