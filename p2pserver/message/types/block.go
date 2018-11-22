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
	"io"

	"github.com/ontio/ontology/common"
	ct "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	comm "github.com/ontio/ontology/p2pserver/common"
)

type Block struct {
	Blk        *ct.Block
	MerkleRoot common.Uint256
}

//Serialize message payload
func (this *Block) Serialization(sink *common.ZeroCopySink) error {
	err := this.Blk.Serialization(sink)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("serialize error. err:%v", err))
	}
	sink.WriteHash(this.MerkleRoot)
	return nil
}

func (this *Block) CmdType() string {
	return comm.BLOCK_TYPE
}

//Deserialize message payload
func (this *Block) Deserialization(source *common.ZeroCopySource) error {
	this.Blk = new(ct.Block)
	err := this.Blk.Deserialization(source)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read Blk error. err:%v", err))
	}

	eof := false
	this.MerkleRoot, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
