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
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

type ShardID struct {
	id uint64
}

const MAX_SHARD_LEVEL = 4
const MAX_CHILD_SHARDS = math.MaxUint16 - 1

var ErrInvalidShardID = errors.New("invalid shard id")

func (self ShardID) Level() int {
	id := self.id
	l1 := id & math.MaxUint16
	l2 := (id >> 16) & math.MaxUint16
	l3 := (id >> 32) & math.MaxUint16
	l4 := (id >> 48) & math.MaxUint16
	lvs := [4]uint64{l4, l3, l2, l1}

	level := MAX_SHARD_LEVEL
	for _, l := range lvs {
		if l != 0 {
			return level
		}
		level -= 1
	}
	return level
}

func (self ShardID) GenSubShardID(index uint16) (ShardID, error) {
	if index == 0 {
		return ShardID{0}, errors.New("wrong child shard index")
	}
	level := self.Level()
	if level == MAX_SHARD_LEVEL {
		return ShardID{0}, errors.New("can not generate sub shard id, max level reached")
	}

	subId := uint64(index) << (16 * uint64(level))

	return ShardID{id: uint64(self.id) + subId}, nil
}

func (self ShardID) Index() uint16 {
	level := self.Level()
	val := self.id >> (uint64(level-1) * 16)
	return uint16(val & math.MaxUint16)
}

func isShardId(id uint64) bool {
	for id&math.MaxUint16 != 0 {
		id >>= 16
	}

	return id == 0
}

func (self ShardID) IsRootShard() bool {
	return self.Level() == 0
}

func NewShardID(id uint64) (ShardID, error) {
	if isShardId(id) == false {
		return ShardID{0}, ErrInvalidShardID
	}

	return ShardID{id}, nil
}

// caller should guarantee the id is valid
func NewShardIDUnchecked(id uint64) ShardID {
	if isShardId(id) == false {
		panic(ErrInvalidShardID)
	}

	return ShardID{id}
}

func (self ShardID) ParentID() ShardID {
	level := self.Level()
	if level == 0 {
		return ShardID{math.MaxUint64}
	}
	id := self.id & ((1 << (uint64(level-1) * 16)) - 1)

	if isShardId(id) == false {
		panic(fmt.Errorf("can not get parent id from %d", self.id))
	}

	return ShardID{id}
}

func (self ShardID) ToUint64() uint64 {
	return self.id
}

// note: (self *ShardID) will not work see: https://github.com/golang/go/issues/7536
func (self ShardID) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.id)
}

func (self *ShardID) UnmarshalJSON(input []byte) error {
	var id uint64
	err := json.Unmarshal(input, &id)
	if err != nil {
		return err
	}
	if isShardId(id) == false {
		return ErrInvalidShardID
	}
	self.id = id
	return nil
}

func ShardIDFromLevels(l1, l2, l3, l4 uint16) (ShardID, error) {
	id := uint64(l1)
	id += uint64(l2) << 16
	id += uint64(l3) << 32
	id += uint64(l4) << 48
	return NewShardID(id)
}
