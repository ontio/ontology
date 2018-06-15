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

func opCat(e *ExecutionEngine) (VMState, error) {
	b2, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	b1, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	r := Concat(b1, b2)
	PushData(e, r)
	return NONE, nil
}

func opSubStr(e *ExecutionEngine) (VMState, error) {
	count, err := PopInt(e)
	if err != nil {
		return FAULT, err
	}
	index, err := PopInt(e)
	if err != nil {
		return FAULT, err
	}
	arr, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	b := arr[index : index+count]
	PushData(e, b)
	return NONE, nil
}

func opLeft(e *ExecutionEngine) (VMState, error) {
	count, err := PopInt(e)
	if err != nil {
		return FAULT, err
	}
	s, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	b := s[:count]
	PushData(e, b)
	return NONE, nil
}

func opRight(e *ExecutionEngine) (VMState, error) {
	count, err := PopInt(e)
	if err != nil {
		return FAULT, err
	}
	arr, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	b := arr[len(arr)-count:]
	PushData(e, b)
	return NONE, nil
}

func opSize(e *ExecutionEngine) (VMState, error) {
	b, err := PopByteArray(e)
	if err != nil {
		return FAULT, err
	}
	PushData(e, len(b))
	return NONE, nil
}
