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

package neovm

import "math/big"

func opInvert(e *ExecutionEngine) (VMState, error) {
	i, err := PopBigInt(e)
	if err != nil {
		return FAULT, err
	}
	PushData(e, big.NewInt(0).Not(i))
	return NONE, nil
}

func opEqual(e *ExecutionEngine) (VMState, error) {
	b1 := PopStackItem(e)
	b2 := PopStackItem(e)
	PushData(e, b1.Equals(b2))
	return NONE, nil
}
