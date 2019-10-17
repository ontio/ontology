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

package snark

import (
	"errors"

	"github.com/kunxian-xia/bn256"
	"github.com/ontio/ontology/common"
)

const fieldSize = 32
const g1Size = 64
const g2Size = 128

// G1 must be encoded in two 32-byte big-endian numbers
func deserializeG1(point *bn256.G1, source *common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g1Size)
	if eof {
		return errors.New("eof when deserialize G1")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G1 point")
	}
	return nil
}

// bn256.G2 is of the form (x, y) where x and y are elements from Fp2
//  and every element from Fp2 is of the form y + x*u
//  Fp2 is isomorphic to Fp[u]/(u^2 - non_residue)
func deserializeG2(point *bn256.G2, source *common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g2Size)
	if eof {
		return errors.New("eof when deserialize G2")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G2 point")
	}
	return nil
}
