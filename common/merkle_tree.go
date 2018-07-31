package common

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
import (
	"crypto/sha256"
)

// param hashes will be used as workspace
func ComputeMerkleRoot(hashes []Uint256) Uint256 {
	if len(hashes) == 0 {
		return Uint256{}
	}
	sha := sha256.New()
	var temp Uint256
	for len(hashes) != 1 {
		n := len(hashes) / 2
		for i := 0; i < n; i++ {
			sha.Reset()
			sha.Write(hashes[2*i][:])
			sha.Write(hashes[2*i+1][:])
			sha.Sum(temp[:0])
			sha.Reset()
			sha.Write(temp[:])
			sha.Sum(hashes[i][:0])
		}
		if len(hashes) == 2*n+1 {
			sha.Reset()
			sha.Write(hashes[2*n][:])
			sha.Write(hashes[2*n][:])

			sha.Sum(temp[:0])
			sha.Reset()
			sha.Write(temp[:])
			sha.Sum(hashes[n][:0])

			hashes = hashes[:n+1]
		} else {
			hashes = hashes[:n]
		}
	}

	return hashes[0]
}
