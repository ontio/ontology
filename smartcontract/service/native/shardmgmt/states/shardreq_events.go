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

package shardstates

import (
	"io"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_REQ_COMMON = iota + 256
)

type CommonShardReq struct {
	SourceShardID uint64 `json:"source_shard_id"`
	Height        uint64 `json:"height"`
	ShardID       uint64 `json:"shard_id"`
	Payload       []byte `json:"payload"`
}

func (evt *CommonShardReq) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *CommonShardReq) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *CommonShardReq) GetHeight() uint64 {
	return evt.Height
}

func (evt *CommonShardReq) GetType() uint32 {
	return EVENT_SHARD_REQ_COMMON
}

func (evt *CommonShardReq) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *CommonShardReq) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}
