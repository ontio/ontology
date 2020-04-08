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

package payload

import (
	"encoding/hex"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestBookkeeper_Serialization(t *testing.T) {
	pubkey, err := hex.DecodeString("039af138392513408f9d1509c651c60066c05b2305de17e44f68088510563e2279")
	assert.Nil(t, err)
	pub, err := keypair.DeserializePublicKey(pubkey)
	assert.Nil(t, err)
	bookkeeper := &Bookkeeper{
		PubKey: pub,
		Action: BookkeeperAction(1),
		Cert:   pubkey,
		Issuer: pub,
	}
	sink := common.NewZeroCopySink(nil)
	bookkeeper.Serialization(sink)
	bookkeeper2 := &Bookkeeper{}
	source := common.NewZeroCopySource(sink.Bytes())
	err = bookkeeper2.Deserialization(source)
	assert.Nil(t, err)

	assert.Equal(t, bookkeeper, bookkeeper2)
}
