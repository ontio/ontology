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

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
)

func SignMsg(account *account.Account, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return signature.Sign(account, data)
}

func hashData(data []byte) common.Uint256 {
	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return common.Uint256(f)
}

func HashMsg(msg ConsensusMsg) (common.Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	return hashData(data), nil
}

type vrfData struct {
	BlockNum          uint32         `json:"block_num"`
	PrevBlockHash     common.Uint256 `json:"prev_block_hash"`
	PrevBlockProposer uint32         `json:"prev_block_proposer"` // TODO: change to NodeID
	TransactionRoot   common.Uint256 `json:"transaction_root"`
	BlockRoot         common.Uint256 `json:"block_root"`
	// TODO: add proposer signature
}

func vrf(block *Block, hash common.Uint256) vconfig.VRFValue {

	data, err := json.Marshal(&vrfData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockHash:     hash,
		PrevBlockProposer: block.getProposer(),
		TransactionRoot:   block.Block.Header.TransactionsRoot,
		BlockRoot:         block.Block.Header.BlockRoot,
	})
	if err != nil {
		return vconfig.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}
