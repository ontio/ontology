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
