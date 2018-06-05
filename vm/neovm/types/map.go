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
	"math/big"

	"github.com/ontio/ontology/vm/neovm/interfaces"

	"reflect"
)

type Map struct {
	_map map[StackItems]StackItems
}

func NewMap() *Map {
	var mp Map
	mp._map = make(map[StackItems]StackItems)
	return &mp
}

func (this *Map) Add(key StackItems, value StackItems) error {
	this._map[key] = value
	return nil
}

func (this *Map) Clear() {
	this._map = make(map[StackItems]StackItems)
}

func (this *Map) ContainsKey(key StackItems) bool {
	_, ok := this._map[key]
	return ok
}

func (this *Map) Remove(key StackItems) {
	delete(this._map, key)
}

func (this *Map) Equals(that StackItems) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Map) GetBoolean() bool {
	return true
}

func (this *Map) GetByteArray() []byte {
	return nil
	//return this.ToArray()
}

func (this *Map) GetBigInteger() *big.Int {
	return nil
}

func (this *Map) GetInterface() interfaces.Interop {
	return nil
}

func (this *Map) GetArray() []StackItems {
	return nil
}

func (this *Map) GetStruct() []StackItems {
	return nil
}

func (this *Map) GetMap() map[StackItems]StackItems {
	return this._map
}

func (this *Map) TryGetValue(key StackItems) StackItems {
	for k, v := range this._map {
		if k.Equals(key) {
			return v
		}
	}
	return nil
	//return this._map[key]
}
