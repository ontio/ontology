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
	"encoding/hex"
	"fmt"
	"hash/crc64"
	"strconv"
	"strings"
)

const NODE_ID_BITS = 64

// NodeID is a unique identifier for each node.
// The node identifier is a marshaled elliptic curve public key.
type NodeID [NODE_ID_BITS / 8]byte

// Bytes returns a byte slice representation of the NodeID
func (n NodeID) Bytes() []byte {
	return n[:]
}

// NodeID prints as a long hexadecimal number.
func (n NodeID) String() string {
	return fmt.Sprintf("%x", n[:])
}

var NilID = NodeID{}

func (n NodeID) IsNil() bool {
	return bytes.Compare(n.Bytes(), NilID.Bytes()) == 0
}

func StringID(in string) (NodeID, error) {
	var id NodeID
	b, err := hex.DecodeString(strings.TrimPrefix(in, "0x"))
	if err != nil {
		return id, err
	} else if len(b) > len(id) {
		return id, fmt.Errorf("wrong length, want %d hex chars", len(b)*2)
	}
	copy(id[:], b)
	return id, nil
}

// ConstructID returns a marshaled representation of the given address:port.
func ConstructID(ip string, port uint16) uint64 {
	var buffer bytes.Buffer
	buffer.WriteString(ip)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(int(port)))

	crcTable := crc64.MakeTable(crc64.ECMA)
	id := crc64.Checksum(buffer.Bytes(), crcTable)
	return id
}
