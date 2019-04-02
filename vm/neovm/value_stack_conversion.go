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

import (
	"github.com/ontio/ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology/vm/neovm/types"
)

func (self *ValueStack) PushBool(val bool) error {
	return self.Push(types.VmValueFromBool(val))
}

func (self *ValueStack) PopAsBool() (bool, error) {
	val, err := self.Pop()
	if err != nil {
		return false, err
	}

	return val.AsBool()
}

func (self *ValueStack) PushInt64(val int64) error {
	return self.Push(types.VmValueFromInt64(val))
}

func (self *ValueStack) PushUint64(val uint64) error {
	return self.Push(types.VmValueFromUint64(val))
}
func (self *ValueStack) PushUint32(val uint32) error {
	return self.Push(types.VmValueFromUint64(uint64(val)))
}

func (self *ValueStack) PopAsInt64() (int64, error) {
	val, err := self.Pop()
	if err != nil {
		return 0, err
	}
	return val.AsInt64()
}

func (self *ValueStack) PopAsIntValue() (types.IntValue, error) {
	val, err := self.Pop()
	if err != nil {
		return types.IntValue{}, err
	}
	return val.AsIntValue()
}

func (self *ValueStack) PushBytes(val []byte) error {
	v, err := types.VmValueFromBytes(val)
	if err != nil {
		return err
	}
	return self.Push(v)
}

func (self *ValueStack) PopAsBytes() ([]byte, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsBytes()
}

func (self *ValueStack) PopAsArray() (*types.ArrayValue, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsArrayValue()
}

func (self *ValueStack) PopAsMap() (*types.MapValue, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsMapValue()
}

func (self *ValueStack) PopAsStruct() (*types.StructValue, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsStructValue()
}

func (self *ValueStack) PushAsInteropValue(val interfaces.Interop) error {
	return self.Push(types.VmValueFromInteropValue(types.NewInteropValue(val)))
}

func (self *ValueStack) PopAsInteropValue() (types.InteropValue, error) {
	val, err := self.Pop()
	if err != nil {
		return types.InteropValue{}, err
	}
	return val.AsInteropValue()
}

func (self *ValueStack) PopPairAsBytes() (left, right []byte, err error) {
	right, err = self.PopAsBytes()
	if err != nil {
		return
	}
	left, err = self.PopAsBytes()
	return
}

func (self *ValueStack) PopPairAsBool() (left, right bool, err error) {
	right, err = self.PopAsBool()
	if err != nil {
		return
	}
	left, err = self.PopAsBool()
	return
}

func (self *ValueStack) PopPairAsInt64() (left, right int64, err error) {
	right, err = self.PopAsInt64()
	if err != nil {
		return
	}
	left, err = self.PopAsInt64()
	return
}

func (self *ValueStack) PopPairAsIntVal() (left, right types.IntValue, err error) {
	right, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	left, err = self.PopAsIntValue()
	return
}

func (self *ValueStack) PopTripleAsBytes() (left, middle, right []byte, err error) {
	right, err = self.PopAsBytes()
	if err != nil {
		return
	}
	middle, err = self.PopAsBytes()
	if err != nil {
		return
	}
	left, err = self.PopAsBytes()
	return
}

func (self *ValueStack) PopTripleAsBool() (left, middle, right bool, err error) {
	right, err = self.PopAsBool()
	if err != nil {
		return
	}
	middle, err = self.PopAsBool()
	if err != nil {
		return
	}
	left, err = self.PopAsBool()
	return
}

func (self *ValueStack) PopTripleAsIntVal() (left, middle, right types.IntValue, err error) {
	right, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	middle, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	left, err = self.PopAsIntValue()
	return
}

func (self *ValueStack) PeekAsBytes(index int64) ([]byte, error) {
	val, err := self.Peek(index)
	if err != nil {
		return nil, err
	}
	bs, err := val.AsBytes()
	if err != nil {
		return nil, err
	}
	return bs, nil
}
