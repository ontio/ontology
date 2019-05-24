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
	comm "github.com/ontio/ontology/p2pserver/common"
)

type BlkHeader struct {
	BlkHdr []*ct.Header
}

type RawBlockHeader struct {
	BlkHdr []*ct.RawHeader
}

func (this *RawBlockHeader) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.BlkHdr)))

	for _, header := range this.BlkHdr {
		header.Serialization(sink)
	}
}
func (this *RawBlockHeader) Deserialization(source *common.ZeroCopySource) error {
	panic("[block_header] unsupport")
}

func (this *RawBlockHeader) CmdType() string {
	return comm.HEADERS_TYPE
}

//Serialize message payload
func (this BlkHeader) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.BlkHdr)))

	for _, header := range this.BlkHdr {
		header.Serialization(sink)
	}
}

func (this *BlkHeader) CmdType() string {
	return comm.HEADERS_TYPE
}

//Deserialize message payload
func (this *BlkHeader) Deserialization(source *common.ZeroCopySource) error {
	var count uint32
	count, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(count); i++ {
		var headers ct.Header
		err := headers.Deserialization(source)
		if err != nil {
			return fmt.Errorf("deserialze BlkHeader error: %v", err)
		}
		this.BlkHdr = append(this.BlkHdr, &headers)
	}
	return nil
}
