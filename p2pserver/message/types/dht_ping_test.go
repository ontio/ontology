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
	"encoding/binary"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func genTestNode() *types.Node {
	node := config.DefConfig.Genesis.DHT.Seeds[0]
	seed := &types.Node{
		IP:      node.IP,
		UDPPort: node.UDPPort,
		TCPPort: node.TCPPort,
	}

	id := types.ConstructID(node.IP, node.UDPPort)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	copy(seed.ID[:], b[:])

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
