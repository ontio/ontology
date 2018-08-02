package types

import (
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func genTestDHTPong() *DHTPong {
	seed := genTestNode()
	addr := &net.UDPAddr{}
	pong := new(DHTPong)
	copy(pong.FromID[:], seed.ID[:])

	pong.SrcEndPoint.UDPPort = seed.UDPPort
	pong.SrcEndPoint.TCPPort = seed.TCPPort

	copy(pong.SrcEndPoint.Addr[:], seed.IP)

	pong.DestEndPoint.UDPPort = uint16(addr.Port)

	destIP := addr.IP.To16()
	copy(pong.DestEndPoint.Addr[:], destIP)
	return pong
}

func TestDHTPong_Serialization(t *testing.T) {
	dhtPong := genTestDHTPong()
	assert.Equal(t, common.DHT_PONG, dhtPong.CmdType())
	bf, err := dhtPong.Serialization()
	assert.Nil(t, err)

	deserializeDhtPong := new(DHTPong)
	err = deserializeDhtPong.Deserialization(bf)
	assert.Nil(t, err)
	assert.Equal(t, dhtPong, deserializeDhtPong)

	MessageTest(t, dhtPong)
}
