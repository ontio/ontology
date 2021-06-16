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

package common

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"sort"
	"time"
)

//type timeHeap []uint64
//
//func (h timeHeap) Len() int           { return len(h) }
//func (h timeHeap) Less(i, j int) bool { return h[i] < h[j] }
//func (h timeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
//
//func (h *timeHeap) Push(x interface{}) {
//	*h = append(*h, x.(uint64))
//}
//
//func (h *timeHeap) Pop() interface{} {
//	old := *h
//	n := len(old)
//	x := old[n-1]
//	*h = old[0 : n-1]
//	return x
//}

type TimedTx struct {
	Time int64
	Tx   *types.Transaction
}

type TxByTime []*TimedTx

func (s TxByTime) Len() int           { return len(s) }
func (s TxByTime) Less(i, j int) bool { return s[i].Time < s[j].Time }
func (s TxByTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type txSortedTimeMap struct {
	items map[common.Uint256]*TimedTx
	//index *timeHeap
	cache TxByTime
}

func newTxSortedTimeMap() *txSortedTimeMap {
	return &txSortedTimeMap{
		items: make(map[common.Uint256]*TimedTx),
	}
}

func (m *txSortedTimeMap) Get(hash common.Uint256) *TimedTx {
	return m.items[hash]
}

func (m *txSortedTimeMap) Put(tx *types.Transaction) {
	if tx.IsEipTx() {
		hash := tx.Hash()
		//if same hash exist, just return
		if m.items[hash] != nil {
			return
		}
		m.items[hash] = &TimedTx{
			Time: time.Now().Unix(),
			Tx:   tx,
		}
		m.cache = nil
	}
}

func (m *txSortedTimeMap) Remove(hash common.Uint256) bool {
	if _, ok := m.items[hash]; !ok {
		return false
	}

	delete(m.items, hash)
	m.cache = nil

	return true
}

func (m *txSortedTimeMap) Len() int {
	return len(m.items)
}

func (m *txSortedTimeMap) flatten() TxByTime {
	// If the sorting was not cached yet, create and cache it
	if m.cache == nil {
		m.cache = make(TxByTime, 0, len(m.items))
		for _, tx := range m.items {
			m.cache = append(m.cache, tx)
		}
		sort.Sort(m.cache)
	}
	return m.cache
}

// LastElement returns the last element of a flattened list, thus, the
// transaction with the highest nonce
func (m *txSortedTimeMap) LastElement() *TimedTx {
	cache := m.flatten()
	return cache[len(cache)-1]
}

func (m *txSortedTimeMap) ExpiredTxByTime(expireTime int64) TxByTime {
	cache := m.flatten()
	expired := make([]*TimedTx, 0)
	for _, c := range cache {
		if c.Time <= expireTime {
			expired = append(expired, c)
		}
	}
	return expired
}
