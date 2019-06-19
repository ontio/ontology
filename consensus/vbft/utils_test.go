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

package vbft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
)

func HashBlock(blk *Block) (common.Uint256, error) {
	return blk.Block.Hash(), nil
}

func TestSignMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg := constructProposalMsgTest(acc)
	_, err := SignMsg(acc, msg)
	if err != nil {
		t.Errorf("TestSignMsg Failed: %v", err)
		return
	}
	t.Log("TestSignMsg succ")
}

func TestHashBlock(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	hash, _ := HashBlock(blk)
	t.Logf("TestHashBlock: %v", hash)
}

func TestHashMsg(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
		return
	}
	blockproposalmsg := &blockProposalMsg{
		Block: blk,
	}
	uint256, err := HashMsg(blockproposalmsg)
	if err != nil {
		t.Errorf("TestHashMsg failed: %v", err)
		return
	}
	t.Logf("TestHashMsg succ: %v\n", uint256)
}

func TestVrfParticipantSeed(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	vrfvalue := getParticipantSelectionSeed(blk)
	if len(vrfvalue) == 0 {
		t.Errorf("TestVrfParticipantSeed failed:")
		return
	}
	t.Log("TestVrfParticipantSeed succ")
}

func TestVrf(t *testing.T) {
	user := account.NewAccount("")
	prevVrf := []byte("test string")
	blkNum := uint32(10)
	v1, p1, err := computeVrf(user.PrivKey(), blkNum, prevVrf)
	if err != nil {
		t.Fatalf("compute vrf: %s", err)
	}

	if err := verifyVrf(user.PubKey(), blkNum, prevVrf, v1, p1); err != nil {
		t.Fatalf("verify vrf: %s", err)
	}

	// test json byte formatting
	data, _ := json.Marshal(&vrfData{10, prevVrf})
	fmt.Println(string(data))
	x := base64.StdEncoding.EncodeToString(prevVrf)
	fmt.Println(x)
}
