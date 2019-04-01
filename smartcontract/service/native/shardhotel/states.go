/*
 * Copyright (C) 2019 The ontology Authors
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

package shardhotel

import (
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"

	"github.com/ontio/ontology/common"
)

type ShardHotelRoomState struct {
	Owner common.Address
}

func (this *ShardHotelRoomState) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("serialize: write owner failed, err: %s", err)
	}
	return nil
}

func (this *ShardHotelRoomState) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner failed, err: %s", err)
	}
	return nil
}

func (this *ShardHotelRoomState) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Owner)
}

func (this *ShardHotelRoomState) Deserialization(source *common.ZeroCopySource) error {
	owner, eof := source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Owner = owner
	return nil
}
