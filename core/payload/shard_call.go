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
package payload

import (
	"github.com/ontio/ontology/core/xshard_types"
	"io"

	"github.com/ontio/ontology/common"
)

type ShardCall struct {
	Msgs []*xshard_types.CommonShardMsg
}

//note: InvokeCode.Code has data reference of param source
func (self *ShardCall) Deserialization(source *common.ZeroCopySource) error {
	n, _, irregular, eof := source.NextVarUint()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := uint64(0); i < n; i++ {
		buf, _, irregular, eof := source.NextVarBytes()
		if irregular {
			return common.ErrIrregularData
		}
		if irregular {
			return common.ErrIrregularData
		}
		if eof {
			return io.ErrUnexpectedEOF
		}
		msg := &xshard_types.CommonShardMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(buf)); err != nil {
			return err
		}

		self.Msgs = append(self.Msgs, msg)
	}

	return nil
}

func (self *ShardCall) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(self.Msgs)))
	for _, msg := range self.Msgs {
		sink.WriteVarBytes(common.SerializeToBytes(msg))
	}
}
