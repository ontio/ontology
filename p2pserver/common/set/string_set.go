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

package set

import (
	"reflect"
)

type empty struct{}

// StringSet is a set of strings, implemented via map[string]struct{} for minimal memory consumption.
type StringSet map[string]empty

// NewStringSet creates a StringSet from a list of values.
func NewStringSet(items ...string) StringSet {
	ss := StringSet{}
	ss.Insert(items...)
	return ss
}

// StringKeySet creates a StringSet from a keys of a map[string](? extends interface{}).
// If the value passed in is not actually a map, this will panic.
func StringKeySet(theMap interface{}) StringSet {
	v := reflect.ValueOf(theMap)
	ret := StringSet{}

	for _, keyValue := range v.MapKeys() {
		ret.Insert(keyValue.Interface().(string))
	}
	return ret
}

// Insert adds items to the set.
func (s StringSet) Insert(items ...string) StringSet {
	for _, item := range items {
		s[item] = empty{}
	}
	return s
}

// Delete removes all items from the set.
func (s StringSet) Delete(items ...string) StringSet {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has returns true if and only if item is contained in the set.
func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// Len returns the size of the set.
func (s StringSet) Len() int {
	return len(s)
}
