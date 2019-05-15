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

import "github.com/ontio/ontology/common"

type CrossShardPayload struct {
	Version uint32
	ShardID common.ShardID
	Data    []byte
}

func (this *CrossShardPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

func (this *CrossShardPayload) Verify() error {
	return nil
}

func (this *CrossShardPayload) Type() common.InventoryType {
	return common.CROSS_SHARD
}
