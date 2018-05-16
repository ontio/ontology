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
)

func TestAddrReqSerializationDeserialization(t *testing.T) {
	var msg AddrReq
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.Init("getaddr", sum, 0)

	buf, err := msg.Serialization()

	var demsg AddrReq
	err = demsg.Deserialization(buf)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("getaddr Test_Deserialization successful")
	}
}
