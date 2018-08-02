package types

import (
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func genTestNode() *types.Node {
	node := config.DefConfig.Genesis.DHT.Seeds[0]
	pubKey, err := hex.DecodeString(node.PubKey)
	k, err := keypair.DeserializePublicKey(pubKey)
	if err != nil {
		return nil
	}
	seed := &types.Node{
		IP:      node.IP,
		UDPPort: node.UDPPort,
		TCPPort: node.TCPPort,
	}
	seed.ID, _ = types.PubkeyID(k)
	return seed
}

func genTestDHTPing() *DHTPing {
	seed := genTestNode()
	addr := &net.UDPAddr{}
	ping := new(DHTPing)
	copy(ping.FromID[:], seed.ID[:])

	ping.SrcEndPoint.UDPPort = seed.UDPPort
	ping.SrcEndPoint.TCPPort = seed.TCPPort

	copy(ping.SrcEndPoint.Addr[:], seed.IP)

	ping.DestEndPoint.UDPPort = uint16(addr.Port)

	destIP := addr.IP.To16()
	copy(ping.DestEndPoint.Addr[:], destIP)
	return ping
}

func TestDHTPing_Serialization(t *testing.T) {
	dhtPing := genTestDHTPing()
	assert.Equal(t, common.DHT_PING, dhtPing.CmdType())

	bf, err := dhtPing.Serialization()
	assert.Nil(t, err)

	deserializeDhtPing := new(DHTPing)
	err = deserializeDhtPing.Deserialization(bf)
	assert.Nil(t, err)
	assert.Equal(t, dhtPing, deserializeDhtPing)

	MessageTest(t, dhtPing)
}
