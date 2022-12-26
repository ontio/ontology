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
//Header index cache
package ledgerstore

import "github.com/ontio/ontology/common"

const (
	HEADER_INDEX_MAX_SIZE = uint32(2000) //Max size of saving header index to cache
)

type HeaderIndexCache struct {
	headerIndex map[uint32]common.Uint256 //Header index, Mapping header height => block hash
	firstIndex  uint32                    //First index of cache
	lastIndex   uint32                    //Last index of cache
}

func NewHeaderIndexCache() *HeaderIndexCache {
	return &HeaderIndexCache{
		headerIndex: make(map[uint32]common.Uint256),
	}
}

func (this *HeaderIndexCache) setHeaderIndex(curBlockHeight uint32, curHeaderHeight uint32, blockHash common.Uint256) {
	this.headerIndex[curHeaderHeight] = blockHash
	if this.getLastIndex() < curHeaderHeight {
		this.lastIndex = curHeaderHeight
	}
	if this.getFirstIndex() < curBlockHeight {
		cacheSize := curBlockHeight - this.getFirstIndex() + 1
		for height := this.getFirstIndex(); cacheSize > HEADER_INDEX_MAX_SIZE; cacheSize-- {
			this.delHeaderIndex(height)
			height++
			this.firstIndex = height
		}
	}
}

func (this *HeaderIndexCache) delHeaderIndex(height uint32) {
	delete(this.headerIndex, height)
}

func (this *HeaderIndexCache) getHeaderIndex(height uint32) common.Uint256 {
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}

func (this *HeaderIndexCache) setLastIndex(lastIndex uint32) {
	this.lastIndex = lastIndex
}

func (this *HeaderIndexCache) getLastIndex() uint32 {
	return this.lastIndex
}

func (this *HeaderIndexCache) setFirstIndex(firstIndex uint32) {
	this.firstIndex = firstIndex
}

func (this *HeaderIndexCache) getFirstIndex() uint32 {
	return this.firstIndex
}

func (this *HeaderIndexCache) getCurrentHeaderHash() common.Uint256 {
	blockHash, ok := this.headerIndex[this.getLastIndex()]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}
