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

package ledgerstore

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestSyncMapRange(t *testing.T) {
	m := sync.Map{}

	for i := 0; i < 10; i++ {
		m.Store("k"+strconv.Itoa(i), "v"+strconv.Itoa(i))
	}
	cnt := 0

	m.Range(func(key, value interface{}) bool {
		fmt.Printf("key :%s, val :%s\n", key, value)

		if key == "k5" {
			return false
		}
		cnt += 1
		return true
	})

	fmt.Println(cnt)

}
