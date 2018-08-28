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
	"crypto/sha256"
	"hash"
)

// checksum implement hash.Hash interface and io.Writer
type checksum struct {
	hash.Hash
}

func (self *checksum) Size() int {
	return CHECKSUM_LEN
}

func (self *checksum) Sum(b []byte) []byte {
	temp := self.Hash.Sum(nil)
	h := sha256.Sum256(temp)

	return append(b, h[:CHECKSUM_LEN]...)
}

func NewChecksum() hash.Hash {
	return &checksum{sha256.New()}
}

func Checksum(data []byte) [CHECKSUM_LEN]byte {
	var checksum [CHECKSUM_LEN]byte
	t := sha256.Sum256(data)
	s := sha256.Sum256(t[:])

	copy(checksum[:], s[:])

	return checksum
}
