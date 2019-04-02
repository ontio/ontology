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
	"github.com/ontio/ontology/vm/neovm/constants"
	"github.com/ontio/ontology/vm/neovm/errors"
)

type ArrayValue struct {
	Data []VmValue
}

const initArraySize = 16

func NewArrayValue() *ArrayValue {
	return &ArrayValue{Data: make([]VmValue, 0, initArraySize)}
}

func (self *ArrayValue) Append(item VmValue) error {
	if len(self.Data) >= constants.MAX_ARRAY_SIZE {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	self.Data = append(self.Data, item)
	return nil
}

func (self *ArrayValue) Len() int64 {
	return int64(len(self.Data))
}

func (self *ArrayValue) RemoveAt(index int64) error {
	if index < 0 || index >= self.Len() {
		return errors.ERR_INDEX_OUT_OF_BOUND
	}
	self.Data = append(self.Data[:index], self.Data[index+1:]...)
	return nil
}
