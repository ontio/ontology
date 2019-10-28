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
package states

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestVoteState_Deserialize_Serialize(t *testing.T) {
	_, pubKey1, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	_, pubKey2, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)

	vs := VoteState{
		StateBase:  StateBase{(byte)(1)},
		PublicKeys: []keypair.PublicKey{pubKey1, pubKey2},
		Count:      10,
	}

	sink := common.NewZeroCopySink(nil)
	vs.Serialization(sink)
	bs := sink.Bytes()

	var vs2 VoteState
	source := common.NewZeroCopySource(bs)
	vs2.Deserialization(source)
	assert.Equal(t, vs, vs2)

	source = common.NewZeroCopySource(bs[:len(bs)-1])
	err := vs2.Deserialization(source)
	assert.NotNil(t, err)
}
