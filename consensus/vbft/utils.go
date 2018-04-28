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
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	. "github.com/Ontology/common"
	"github.com/Ontology/consensus/vbft/config"
	"github.com/Ontology/crypto"
)

func SignMsg(sk []byte, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return crypto.Sign(sk, data)
}

func HashBlock(blk *Block) (Uint256, error) {
	return blk.Block.Hash(), nil
}

func HashMsg(msg ConsensusMsg) (Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return Uint256(f), nil
}

type vrfData struct {
	BlockNum          uint64  `json:"block_num"`
	PrevBlockHash     Uint256 `json:"prev_block_hash"`
	PrevBlockProposer uint32  `json:"prev_block_proposer"` // TODO: change to NodeID
	PrevBlockSig      []byte  `json:"prev_block_sig"`
}

func vrf(block *Block, hash Uint256) vconfig.VRFValue {

	// XXX: all-zero vrf value is taken as invalid
	sig := []byte{}
	if len(block.Block.Header.SigData) > 0 {
		sig = block.Block.Header.SigData[0]
	}
	if block.isEmpty() {
		sig = block.Block.Header.SigData[1]
	}

	data, err := json.Marshal(&vrfData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockHash:     hash,
		PrevBlockProposer: block.getProposer(),
		PrevBlockSig:      sig,
	})
	if err != nil {
		return vconfig.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}
