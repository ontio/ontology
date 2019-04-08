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

package shardccmc

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegisterCCParam struct {
	ShardID      types.ShardID
	Owner        common.Address
	ContractAddr common.Address
	Dependencies []common.Address
}

func (this *RegisterCCParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shardId failed, err: %s", err)
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
	for index, dep := range this.Dependencies {
		if err := utils.WriteAddress(w, dep); err != nil {
			return fmt.Errorf("serialize: write dependencies failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *RegisterCCParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shardId failed, err: %s", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner failed, err: %s", err)
	}
	if this.ContractAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read contract owner failed, err: %s", err)
	}
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read dependencies num failed, err: %s", err)
	}
	this.Dependencies = make([]common.Address, num)
	for i := uint64(0); i < num; i++ {
		addr, err := utils.ReadAddress(r)
		if err != nil {
			return fmt.Errorf("deserialize: read dependencies failed, index %d, err: %s", i, err)
		}
		this.Dependencies[i] = addr
	}
	return nil
}
