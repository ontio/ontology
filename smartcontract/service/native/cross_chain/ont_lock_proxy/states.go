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

package ont_lock_proxy

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// Args for lock and unlock
type Args struct {
	ToAddress []byte
	Value     uint64
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.ToAddress)
	utils.EncodeVarUint(sink, this.Value)

}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ToAddress, err = utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("Args.Deserialization DecodeVarBytes error:%s", err)
	}
	this.Value, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("Args.Deserialization DecodeVarUint error:%s", err)
	}
	return nil
}

type LockParam struct {
	ToChainID   uint64
	FromAddress common.Address
	Fee         uint64
	Args        Args
}

func (this *LockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeAddress(sink, this.FromAddress)
	utils.EncodeVarUint(sink, this.Fee)
	this.Args.Serialization(sink)
}

func (this *LockParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.ToChainID, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeVarUint error:%s", err)
	}
	this.FromAddress, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeAddress error:%s", err)
	}
	this.Fee, err = utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("LockParam.Deserialization DecodeAddress error:%s", err)
	}
	err = this.Args.Deserialization(source)
	if err != nil {
		return err
	}
	return nil
}
