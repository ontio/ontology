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
package auth

import "testing"

//{"a", "b"} == {"b", "a"}
func testEq(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}
	Map := make(map[string]bool)
	for i := range a {
		Map[a[i]] = true
	}
	for _, s := range b {
		_, ok := Map[s]
		if !ok {
			return false
		}
	}
	return true
}
func TestStringSliceUniq(t *testing.T) {
	s := []string{"foo", "foo1", "foo2", "foo", "foo1", "foo2", "foo3"}
	ret := stringSliceUniq(s)
	t.Log(ret)
	if !testEq(ret, []string{"foo", "foo1", "foo2", "foo3"}) {
		t.Fatalf("failed")
	}
}
