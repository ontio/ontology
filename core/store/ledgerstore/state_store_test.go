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
	"math/rand"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/merkle"
	"github.com/stretchr/testify/assert"
)

func TestContractState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	testCode := []byte("testcode")

	deploy := &payload.DeployCode{
		Code:        testCode,
		NeedStorage: false,
		Name:        "testsm",
		Version:     "v1.0",
		Author:      "",
		Email:       "",
		Description: "",
	}

	address := common.AddressFromVmCode(testCode)
	err = batch.TryGetOrAdd(
		scommon.ST_CONTRACT,
		address[:],
		deploy)
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
	contractState1, err := testStateStore.GetContractState(address)
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

func TestBookkeeperState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")

	currBookkeepers := make([]keypair.PublicKey, 0)
	currBookkeepers = append(currBookkeepers, acc1.PublicKey)
	currBookkeepers = append(currBookkeepers, acc2.PublicKey)
	nextBookkeepers := make([]keypair.PublicKey, 0)
	nextBookkeepers = append(nextBookkeepers, acc1.PublicKey)
	nextBookkeepers = append(nextBookkeepers, acc2.PublicKey)

	bookkeeperState := &states.BookkeeperState{
		CurrBookkeeper: currBookkeepers,
		NextBookkeeper: nextBookkeepers,
	}

	batch.TryAdd(scommon.ST_BOOKKEEPER, BOOKKEEPER, bookkeeperState)
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
	bookState, err := testStateStore.GetBookkeeperState()
	if err != nil {
		t.Errorf("GetBookkeeperState error %s", err)
		return
	}
	currBookkeepers1 := bookState.CurrBookkeeper
	nextBookkeepers1 := bookState.NextBookkeeper
	for index, pk := range currBookkeepers {
		pk1 := currBookkeepers1[index]
		if !keypair.ComparePublicKey(pk, pk1) {
			t.Errorf("TestBookkeeperState currentBookkeeper failed")
			return
		}
	}
	for index, pk := range nextBookkeepers {
		pk1 := nextBookkeepers1[index]
		if !keypair.ComparePublicKey(pk, pk1) {
			t.Errorf("TestBookkeeperState nextBookkeeper failed")
			return
		}
	}
}

func getStateBatch() (*statestore.StateBatch, error) {
	testStateStore.NewBatch()
	batch := testStateStore.NewStateBatch()
	return batch, nil
}

func TestStateMerkleRoot(t *testing.T) {
	teststatemerkleroot := func(H, effectiveStateHashHeight uint32) {
		diffHashes := make([]common.Uint256, 0, H)
		for i := uint32(0); i < H; i++ {
			var hash common.Uint256
			rand.Read(hash[:])
			diffHashes = append(diffHashes, hash)
		}
		db := NewMemStateStore(effectiveStateHashHeight)
		for h, hash := range diffHashes[:effectiveStateHashHeight] {
			height := uint32(h)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root, _ := db.GetStateMerkleRoot(height)
			assert.Equal(t, root, common.UINT256_EMPTY)
		}

		merkleTree := merkle.NewTree(0, nil, nil)
		for h, hash := range diffHashes[effectiveStateHashHeight:] {
			height := uint32(h) + effectiveStateHashHeight
			merkleTree.AppendHash(hash)
			root1 := db.GetStateMerkleRootWithNewHash(hash)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root2, _ := db.GetStateMerkleRoot(height)
			root3 := merkleTree.Root()

			assert.Equal(t, root1, root2)
			assert.Equal(t, root1, root3)
		}
	}

	for i := 0; i < 200; i++ {
		teststatemerkleroot(1024, uint32(i))
		h := rand.Uint32()%1000 + 1
		eff := rand.Uint32() % h
		teststatemerkleroot(h, eff)
	}

}
