/*
 * Copyright (C) 2019 The ontology Authors
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

package shardping

import (
	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
)

type ShardPingParam struct {
	FromShard types.ShardID
	ToShard   types.ShardID
	Param     string
}

func (this *ShardPingParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.FromShard); err != nil {
		return fmt.Errorf("serialize: write from shard failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ToShard); err != nil {
		return fmt.Errorf("serialize: write to shard failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.Param); err != nil {
		return fmt.Errorf("serialize: write param failed, err: %s", err)
	}
	return nil
}

func (this *ShardPingParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.FromShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read from shard failed, err: %s", err)
	}
	if this.ToShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	if this.Param, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read param failed, err: %s", err)
	}
	return nil
}
