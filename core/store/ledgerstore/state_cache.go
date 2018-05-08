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
	"github.com/hashicorp/golang-lru"
	"github.com/ontio/ontology/core/states"
)

const (
	STATE_CACHE_SIZE = 100000
)

type StateCache struct {
	stateCache *lru.ARCCache
}

func NewStateCache() (*StateCache, error) {
	stateCache, err := lru.NewARC(STATE_CACHE_SIZE)
	if err != nil {
		return nil, err
	}
	return &StateCache{
		stateCache: stateCache,
	}, nil
}

func (this *StateCache) GetState(key []byte) states.StateValue {
	state, ok := this.stateCache.Get(string(key))
	if !ok {
		return nil
	}
	return state.(states.StateValue)
}

func (this *StateCache) AddState(key []byte, state states.StateValue) {
	this.stateCache.Add(string(key), state)
}

func (this *StateCache) DeleteState(key []byte) {
	this.stateCache.Remove(string(key))
}
