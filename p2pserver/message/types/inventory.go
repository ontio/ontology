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
	"io"

	"github.com/ontio/ontology/common"
	p2pCommon "github.com/ontio/ontology/p2pserver/common"
)

var LastInvHash common.Uint256

type InvPayload struct {
	InvType common.InventoryType
	Blk     []common.Uint256
}

type Inv struct {
	P InvPayload
}

func (this Inv) invType() common.InventoryType {
	return this.P.InvType
}

func (this *Inv) CmdType() string {
	return p2pCommon.INV_TYPE
}

//Serialize message payload
func (this Inv) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint8(uint8(this.P.InvType))

	blkCnt := uint32(len(this.P.Blk))
	sink.WriteUint32(blkCnt)
	for _, hash := range this.P.Blk {
		sink.WriteHash(hash)
	}

	return nil
}

//Deserialize message payload
func (this *Inv) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	invType, eof := source.NextUint8()
	this.P.InvType = common.InventoryType(invType)
	blkCnt, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(blkCnt); i++ {
		hash, eof := source.NextHash()
		if eof {
			return io.ErrUnexpectedEOF
		}

		this.P.Blk = append(this.P.Blk, hash)
	}

	if blkCnt > p2pCommon.MAX_INV_BLK_CNT {
		blkCnt = p2pCommon.MAX_INV_BLK_CNT
	}
	this.P.Blk = this.P.Blk[:blkCnt]
	return nil
}
