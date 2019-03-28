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

package shardutil

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	"io"

	"github.com/ontio/ontology/common/serialization"
)

func GetUint64Bytes(num uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint64(bf, num); err != nil {
		return nil, fmt.Errorf("serialization.WriteUint64, serialize uint64 error: %v", err)
	}
	return bf.Bytes(), nil
}

func GetBytesUint64(b []byte) (uint64, error) {
	num, err := serialization.ReadUint64(bytes.NewBuffer(b))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint64, deserialize uint64 error: %v", err)
	}
	return num, nil
}

func GetUint32Bytes(num uint32) []byte {
	sink := common.NewZeroCopySink(4)
	sink.WriteUint32(num)
	return sink.Bytes()
}

func GetBytesUint32(b []byte) (uint32, error) {
	source := common.NewZeroCopySource(b)
	num, eof := source.NextUint32()
	if eof {
		return 0, io.ErrUnexpectedEOF
	}
	return num, nil
}
