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

package message

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

const (
	ShardGetGenesisBlockReq = iota
	ShardGetGenesisBlockRsp
	ShardGetPeerInfoReq
	ShardGetPeerInfoRsp
)

type ShardSystemEventMsg struct {
	FromAddress common.Address   `json:"from_address"`
	Event       *ShardEventState `json:"event"`
}

type ShardEventState struct {
	Version    uint32        `json:"version"`
	EventType  uint32        `json:"event_type"`
	ToShard    types.ShardID `json:"to_shard"`
	FromHeight uint32        `json:"from_height"`
	Payload    []byte        `json:"payload"`
}
