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

package ledgerstore

import (
	"fmt"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/crypto"
	vmtypes"github.com/Ontology/vm/types"
	"testing"
)

func init() {
	crypto.SetAlg("")
}

func TestContractState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	testCode := []byte("testcode")

	deploy := &payload.DeployCode{
		Code:        testCode,
		VmType:      vmtypes.NEOVM,
		NeedStorage: false,
		Name:        "testsm",
		Version:     "v1.0",
		Author:     "",
		Email:       "",
		Description: "",
	}
	code := &vmtypes.VmCode{
		Code:   testCode,
		VmType: vmtypes.NEOVM,
	}
	codeHash := code.AddressFromVmCode()
	err = batch.TryGetOrAdd(
		scommon.ST_Contract,
		codeHash[:],
		deploy,
		false)
	if err != nil {
		 t.Errorf("TryGetOrAdd contract error %s", err)
		 return
	}

	err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	contractState1, err := testStateStore.GetContractState(codeHash)
	if err != nil {
		t.Errorf("GetContractState error %s", err)
		return
	}
	if contractState1.Name != deploy.Name ||
		contractState1.Version != deploy.Version ||
		contractState1.Author != deploy.Author ||
		contractState1.Description != deploy.Description ||
		contractState1.Email != deploy.Email {
		t.Errorf("TestContractState failed %+v != %+v", contractState1, deploy)
		return
	}
}

func TestBookKeeperState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}

	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	currBookKeepers := make([]*crypto.PubKey, 0)
	currBookKeepers = append(currBookKeepers, &pubKey1)
	currBookKeepers = append(currBookKeepers, &pubKey2)
	nextBookKeepers := make([]*crypto.PubKey, 0)
	nextBookKeepers = append(nextBookKeepers, &pubKey1)
	nextBookKeepers = append(nextBookKeepers, &pubKey2)

	bookKeeperState := &states.BookKeeperState{
		CurrBookKeeper: currBookKeepers,
		NextBookKeeper: nextBookKeepers,
	}
	batch.TryAdd(scommon.ST_BookKeeper, BookerKeeper, bookKeeperState, false)
	err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	bookState, err := testStateStore.GetBookKeeperState()
	if err != nil {
		t.Errorf("GetBookKeeperState error %s", err)
		return
	}
	currBookKeepers1 := bookState.CurrBookKeeper
	nextBookKeepers1 := bookState.NextBookKeeper
	for index, pk := range currBookKeepers {
		pk1 := currBookKeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookKeeperState currentBookKeeper failed")
			return
		}
	}
	for index, pk := range nextBookKeepers {
		pk1 := nextBookKeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookKeeperState currentBookKeeper failed")
			return
		}
	}
}

func getStateBatch() (*statestore.StateBatch, error) {
	err := testStateStore.NewBatch()
	if err != nil {
		return nil, fmt.Errorf("testStateStore.NewBatch error %s", err)
	}
	batch := testStateStore.NewStateBatch()
	return batch, nil
}
