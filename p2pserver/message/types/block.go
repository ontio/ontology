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

package types

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	comm "github.com/ontio/ontology/p2pserver/common"
)

type Block struct {
	Blk        *types.Block
	MerkleRoot common.Uint256
	CCMsg      *types.CrossChainMsg
}

//Serialize message payload
func (this *Block) Serialization(sink *common.ZeroCopySink) {
	this.Blk.Serialization(sink)
	sink.WriteHash(this.MerkleRoot)
	sink.WriteBool(this.CCMsg != nil)
	if this.CCMsg != nil {
		this.CCMsg.Serialization(sink)
	}
}

func (this *Block) CmdType() string {
	return comm.BLOCK_TYPE
}

//Deserialize message payload
func (this *Block) Deserialization(source *common.ZeroCopySource) error {
	this.Blk = new(types.Block)
	err := this.Blk.Deserialization(source)
	if err != nil {
		return fmt.Errorf("read Blk error. err:%v", err)
	}
	var eof bool
	this.MerkleRoot, eof = source.NextHash()
	if eof {
		// to accept old node's block
		this.MerkleRoot = common.UINT256_EMPTY
	}
	hasCCM, irr, eof := source.NextBool()
	if irr || eof {
		// to accept old node's cross msg
		return nil
	}
	var ccMsg *types.CrossChainMsg
	if hasCCM {
		ccMsg = new(types.CrossChainMsg)
		if err := ccMsg.Deserialization(source); err != nil {
			return err
		}
	}
	this.CCMsg = ccMsg

	return nil
}
