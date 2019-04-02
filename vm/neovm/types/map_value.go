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
	"sort"
)

type MapValue struct {
	Data map[string][2]VmValue
}

func NewMapValue() *MapValue {
	return &MapValue{Data: make(map[string][2]VmValue)}
}
func (this *MapValue) Set(key, value VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	this.Data[skey] = [2]VmValue{key, value}
	return nil
}

func (this *MapValue) Reset() {
	this.Data = make(map[string][2]VmValue)
}

func (this *MapValue) Remove(key VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	delete(this.Data, skey)

	return nil
}

func (this *MapValue) Get(key VmValue) (value VmValue, ok bool, err error) {
	skey, e := key.GetMapKey()
	if e != nil {
		err = e
		return
	}

	val, ok := this.Data[skey]
	value = val[1]
	return
}

func (this *MapValue) GetMapSortedKey() []VmValue {
	sortedKeys := this.getMapSortedKey()
	sortedKey := make([]VmValue, 0, len(sortedKeys))
	for _, k := range sortedKeys {
		sortedKey = append(sortedKey, this.Data[k][0])
	}
	return sortedKey
}

func (this *MapValue) getMapSortedKey() []string {
	var unsortKey []string
	for k := range this.Data {
		unsortKey = append(unsortKey, k)
	}
	sort.Strings(unsortKey)
	return unsortKey
}

func (this *MapValue) GetValues() ([]VmValue, error) {
	keys := this.getMapSortedKey()
	values := make([]VmValue, 0, len(this.Data))
	for _, v := range keys {
		values = append(values, this.Data[v][1])
	}
	return values, nil
}
