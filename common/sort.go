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
package common

import (
	"bytes"
	"sort"
)

type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SortUint32s(l []uint32) {
	sort.Stable(Uint32Slice(l))
}

type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SortUint64s(l []uint64) {
	sort.Stable(Uint64Slice(l))
}

type AddressSlice []Address

func (p AddressSlice) Len() int { return len(p) }
func (p AddressSlice) Less(i, j int) bool {
	return bytes.Compare(p[i][:], p[j][:]) < 0
}
func (p AddressSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func SortAddress(l []Address) {
	sort.Stable(AddressSlice(l))
}
