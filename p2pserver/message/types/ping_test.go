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
	"testing"

	"github.com/ontio/ontology/common"
)

func TestPingSerializationDeserialization(t *testing.T) {
	var msg Ping
	msg.Height = make(map[uint64]*HeightInfo)
	heightInfo := &HeightInfo{
		Height:  uint32(1),
		MsgHash: common.Uint256{1, 2, 3},
	}
	msg.Height[0] = heightInfo

	MessageTest(t, &msg
}
