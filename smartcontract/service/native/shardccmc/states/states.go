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
	"fmt"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"

	"github.com/ontio/ontology/common"
)

type ShardCCMCState struct {
	NextCCID uint64
}

func (this *ShardCCMCState) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.NextCCID); err != nil {
		return fmt.Errorf("serialize: write next ccid failed, err: %s", err)
	}
	return nil
}

func (this *ShardCCMCState) Deserialize(r io.Reader) error {
	ccid, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read next ccid fialed, err: %s", err)
	}
	this.NextCCID = ccid
	return nil
}

type ShardCCInfo struct {
	CCID         uint64
	ShardID      types.ShardID
	Owner        common.Address
	ContractAddr common.Address
	Dependencies []common.Address
}

func (this *ShardCCInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.CCID); err != nil {
		return fmt.Errorf("serialize: write ccid failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("serialize: write owner failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.ContractAddr); err != nil {
		return fmt.Errorf("serialize: write contract addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Dependencies))); err != nil {
		return fmt.Errorf("serialize: write dependencies num failed, err: %s", err)
	}
	for i, dep := range this.Dependencies {
		if err := utils.WriteAddress(w, dep); err != nil {
			return fmt.Errorf("serialize: write dependencies failed, index %d, err: %s", i, err)
		}
	}
	return nil
}

func (this *ShardCCInfo) Deserialize(r io.Reader) error {
	var err error = nil
	if this.CCID, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read ccid failed, err: %s", err)
	}
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner failed, err: %s", err)
	}
	if this.ContractAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read contract owner failed, err: %s", err)
	}
	depNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read dependencies num failed, err: %s", err)
	}
	this.Dependencies = make([]common.Address, depNum)
	for i := uint64(0); i < depNum; i++ {
		dep, err := utils.ReadAddress(r)
		if err != nil {
			return fmt.Errorf("deserialize: read dependencies failed, index %d, err: %s", i, err)
		}
		this.Dependencies[i] = dep
	}
	return nil
}
