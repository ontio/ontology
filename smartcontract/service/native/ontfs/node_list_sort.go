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
	"encoding/binary"

	"github.com/ontio/ontology/common"
	"golang.org/x/crypto/sha3"
)

func genRandSlice(listLen uint64, seed []byte, addr [20]byte) []uint16 {
	uint16Slice := make([]uint16, listLen)

	h := sha3.New512()
	h.Write(addr[:])
	h.Write(seed)
	finalRandData := h.Sum(nil)

	for {
		if uint64(len(finalRandData)) >= 2*listLen {
			break
		}
		h.Reset()
		h.Write(addr[:])
		h.Write(finalRandData)
		randData := h.Sum(nil)

		finalRandData = append(finalRandData, randData...)
	}
	finalRandData = finalRandData[0 : 2*listLen]

	for i := uint64(0); i < listLen; i++ {
		uint16Slice[i] = binary.LittleEndian.Uint16(finalRandData[i*2 : (i+1)*2])
	}
	return uint16Slice
}

func sortByRandSlice(values []uint16, nodeAddrList []common.Address) []common.Address {
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
				nodeAddrList[i], nodeAddrList[j] = nodeAddrList[j], nodeAddrList[i]
			}
		}
	}
	return nodeAddrList
}
