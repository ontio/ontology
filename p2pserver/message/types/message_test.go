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
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/stretchr/testify/assert"
)

func TestMsgHdrSerializationDeserialization(t *testing.T) {
	hdr := newMessageHeader("hdrtest", 0, common.Checksum(nil))

	sink := common2.NewZeroCopySink(nil)
	writeMessageHeaderInto(sink, hdr)

	dehdr, err := readMessageHeader(bytes.NewBuffer(sink.Bytes()))
	assert.Nil(t, err)

	assert.Equal(t, hdr, dehdr)

}

func readMessageHeader_old(reader io.Reader) (messageHeader, error) {
	msgh := messageHeader{}
	err := binary.Read(reader, binary.LittleEndian, &msgh)
	return msgh, err
}

func TestMsgHdr2(t *testing.T) {
	hdr := newMessageHeader("hdrtest1", 20, common.Checksum([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	sink := common2.NewZeroCopySink(nil)
	writeMessageHeaderInto(sink, hdr)

	hdr1, err := readMessageHeader(bytes.NewBuffer(sink.Bytes()))
	if err != nil {
		t.Fatalf("read hdr: %s", err)
	}
	if hdr1.Length != hdr.Length ||
		hdr1.Magic != hdr.Magic ||
		hdr1.CMD != hdr.CMD ||
		hdr1.Checksum != hdr.Checksum {
		t.Fatalf("invalid hdr1: %v", hdr1)
	}
}

func BenchmarkReadMessageUseSink(b *testing.B) {
	b.StopTimer()
	hdr := newMessageHeader("hdrtest2", 20, common.Checksum([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	sink := common2.NewZeroCopySink(nil)
	writeMessageHeaderInto(sink, hdr)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		readMessageHeader(bytes.NewBuffer(sink.Bytes()))
	}
}

func BenchmarkReadMessageBinaryRead(b *testing.B) {
	b.StopTimer()
	hdr := newMessageHeader("hdrtest2", 20, common.Checksum([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))

	sink := common2.NewZeroCopySink(nil)
	writeMessageHeaderInto(sink, hdr)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		readMessageHeader_old(bytes.NewBuffer(sink.Bytes()))
	}
}
