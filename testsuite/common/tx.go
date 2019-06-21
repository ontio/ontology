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

package TestCommon

import (
	"fmt"
	bcomm "github.com/ontio/ontology/http/base/common"
	"math"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	cutils "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/types"
)

func CreateAdminTx(t *testing.T, shard common.ShardID, gasPrice uint64, addr common.Address, method string,
	args []interface{}) *types.Transaction {
	if args == nil {
		args = make([]interface{}, 0)
	}
	if len(args) == 0 {
		args = append(args, "")
	}
	mutable, err := bcomm.NewNativeInvokeTransaction(gasPrice, math.MaxUint64, addr, byte(1), method, args)
	if err != nil {
		t.Fatalf("generate tx failed, err: %s", err)
	}

	shardName := chainmgr.GetShardName(shard)
	pks := make([]keypair.PublicKey, 0)
	accounts := make([]*account.Account, 0)
	for i := 0; i < 7; i++ {
		acc := GetAccount(shardName + "_peerOwner" + fmt.Sprintf("%d", i))
		pks = append(pks, acc.PublicKey)
		accounts = append(accounts, acc)
	}

	for _, acc := range accounts {
		if err := cutils.MultiSigTransaction(mutable, 5, pks, acc); err != nil {
			t.Fatalf("multi sign tx: %s", err)
		}
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		t.Fatalf("to immutable tx: %s", err)
	}

	return tx
}

func CreateNativeTx(t *testing.T, user string, gasPrice uint64, addr common.Address, method string,
	args []interface{}) *types.Transaction {
	if args == nil {
		args = make([]interface{}, 0)
	}
	if len(args) == 0 {
		args = append(args, "")
	}
	acc := GetAccount(user)
	if acc == nil {
		t.Fatalf("Invalid user: %s", user)
	}

	mutable, err := bcomm.NewNativeInvokeTransaction(gasPrice, math.MaxUint64, addr, byte(1), method, args)
	if err != nil {
		t.Fatalf("generate tx failed, err: %s", err)
	}
	err = cutils.SignTransaction(acc, mutable)
	if err != nil {
		t.Fatalf("sign tx failed, err: %s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		t.Fatalf("to immutable tx: %s", err)
	}
	return tx
}

func CreateNeoInvokeTx(t *testing.T, user string, addr common.Address, params []interface{}) *types.Transaction {
	return nil
}

func CreateNeoDeployTx(t *testing.T, user string, code []byte, contractName string) *types.Transaction {
	return nil
}
