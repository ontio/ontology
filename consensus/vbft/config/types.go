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

package vconfig

import (
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

// PubkeyID returns a marshaled representation of the given public key.
func PubkeyID(pub keypair.PublicKey) string {
	return common.PubKeyToHex(pub)
}

func Pubkey(nodeid string) (keypair.PublicKey, error) {
	return common.PubKeyFromHex(nodeid)
}

func MustPubkey(nodeid string) keypair.PublicKey {
	key, err := common.PubKeyFromHex(nodeid)
	if err != nil {
		panic(fmt.Errorf("pubkey ID must be valid: got %s", nodeid))
	}
	return key
}

func VbftBlock(header *types.Header) (*VbftBlockInfo, error) {
	blkInfo := &VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}
	return blkInfo, nil
}
