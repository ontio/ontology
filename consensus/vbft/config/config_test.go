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

package vconfig

import (
	"bytes"
	"fmt"
	"testing"
)

func generTestData() []byte {
	nodeId, _ := StringID("206520e7475798520164487f7e4586bb55790097ceb786aab6d5bc889d12991a5a204c6298bef1bf43c20680a3979a213392b99c97042ebae27d2a7af6442aa7c008")
	chainPeers := make([]*PeerConfig, 0)
	peerconfig := &PeerConfig{
		Index: 12,
		ID:    nodeId,
	}
	chainPeers = append(chainPeers, peerconfig)

	tests := &ChainConfig{
		Version:       1,
		View:          12,
		N:             4,
		F:             3,
		BlockMsgDelay: 1000,
		HashMsgDelay:  1000,
		PeerHandshakeTimeout:10000,
		Peers:         chainPeers,
		PosTable:      []uint32{2, 3, 1, 3, 1, 3, 2, 3, 2, 3, 2, 1, 3},
	}
	cc := new(bytes.Buffer)
	tests.Serialize(cc)
	return cc.Bytes()
}
func TestSerialize(t *testing.T) {
	res := generTestData()
	fmt.Println("serialize:", res)
}

func TestDeserialize(t *testing.T) {
	res := generTestData()
	test := &ChainConfig{}
	err := test.Deserialize(bytes.NewReader(res))

	if err != nil {
		t.Log("test failed ")
	}
	fmt.Printf("version: %d, F:%d \n", test.Version, test.F)
	fmt.Println("peers:      ", test.Peers[0].ID.String())
	fmt.Println("dpostable:", test.PosTable[0])
}
