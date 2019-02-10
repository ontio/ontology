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

package ccmc_states

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

type ShardCCMCState struct {
	NextCCID uint64 `json:"next_ccid"`
}

func (this *ShardCCMCState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardCCMCState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardCCInfo struct {
	CCID         uint64           `json:"ccid"`
	ShardID      uint64           `json:"shard_id"`
	Owner        common.Address   `json:"owner"`
	ContractAddr common.Address   `json:"contract_addr"`
	Dependencies []common.Address `json:"dependencies"`
}

func (this *ShardCCInfo) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardCCInfo) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
