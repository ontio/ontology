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

package shardgas

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
)

func TestPeerWithdrawGasParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &PeerWithdrawGasParam{
		Signer:     acc.Address,
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		User:       acc.Address,
		ShardId:    0,
		Amount:     10000,
		WithdrawId: 1,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	newParam := &PeerWithdrawGasParam{}
	err = newParam.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, param, newParam)
}
