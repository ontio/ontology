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
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/header_sync"
	"github.com/stretchr/testify/assert"
)

func TestPeer(t *testing.T) {
	peer := header_sync.Peer{
		Index:      1,
		PeerPubkey: "testPubkey",
	}
	sink := common.NewZeroCopySink(nil)
	peer.Serialization(sink)

	var p header_sync.Peer
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, peer, p)
}

func TestKeyHeights(t *testing.T) {
	key := header_sync.KeyHeights{
		HeightList: []uint32{1, 2, 3, 4},
	}
	sink := common.NewZeroCopySink(nil)
	key.Serialization(sink)

	var k header_sync.KeyHeights
	err := k.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, key, k)
}

func TestConsensusPeers(t *testing.T) {
	peers := header_sync.ConsensusPeers{
		ChainID: 1,
		Height:  2,
		PeerMap: map[string]*header_sync.Peer{
			"testPubkey1": {
				Index:      1,
				PeerPubkey: "testPubkey1",
			},
			"testPubkey2": {
				Index:      2,
				PeerPubkey: "testPubkey2",
			},
		},
	}
	sink := common.NewZeroCopySink(nil)
	peers.Serialization(sink)

	var p header_sync.ConsensusPeers
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, p, peers)
}
