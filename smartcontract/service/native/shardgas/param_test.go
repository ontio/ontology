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
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
)

func TestPeerWithdrawGasParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &PeerWithdrawGasParam{
		Signer:     acc.Address,
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		User:       acc.Address,
		ShardId:    common.NewShardIDUnchecked(0),
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

func TestDepositGasParamSerialize(t *testing.T) {
	user, _ := common.AddressFromBase58("ARpjnrnHEjXhg4aw7vY6xsY6CfQ1XEWzWC")
	param := &DepositGasParam{
		User:    user,
		Amount:  1000000000,
		ShardId: common.NewShardIDUnchecked(1),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	newParam := &DepositGasParam{}
	err = newParam.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, param, newParam)
}
