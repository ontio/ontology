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

package ontfs

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestFsNodeInfo_Serialization(t *testing.T) {
	nodeInfo := FsNodeInfo{
		Pledge:      uint64(10),
		Profit:      uint64(20),
		Volume:      uint64(30),
		RestVol:     uint64(40),
		ServiceTime: uint64(50),
		NodeAddr: common.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05,
			0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05},
		NodeNetAddr: []byte("111.111.111.111：111"),
	}
	sink := common.NewZeroCopySink(nil)
	nodeInfo.Serialization(sink)

	fmt.Printf("%v", sink.Bytes())

	nodeInfo2 := FsNodeInfo{}
	src := common.NewZeroCopySource(sink.Bytes())
	if err := nodeInfo2.Deserialization(src); err != nil {
		t.Fatal("nodeInfo2 deserialize fail!", err.Error())
	}

	assert.Equal(t, nodeInfo, nodeInfo2)
}
