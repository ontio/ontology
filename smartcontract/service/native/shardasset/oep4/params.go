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
package oep4

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegisterParam struct {
	Name        string
	Symbol      string
	Decimals    uint64
	TotalSupply *big.Int
	Account     common.Address // receive total supply asset while init
}

func (this *RegisterParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Name); err != nil {
		return fmt.Errorf("serialize: write name failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.Symbol); err != nil {
		return fmt.Errorf("serialize: write name failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Decimals); err != nil {
		return fmt.Errorf("serialize: write decimals failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.TotalSupply)); err != nil {
		return fmt.Errorf("serialize: write total supply failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Account); err != nil {
		return fmt.Errorf("serialize: write account failed, err: %s", err)
	}
	return nil
}

func (this *RegisterParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Name, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read name failed, err: %s", err)
	}
	if this.Symbol, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read symbol failed, err: %s", err)
	}
	if this.Decimals, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read decimals failed, err: %s", err)
	}
	if supply, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read total supply failed, err: %s", err)
	} else {
		this.TotalSupply = common.BigIntFromNeoBytes(supply)
	}
	if this.Account, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read account failed, err: %s", err)
	}
	return nil
}

type MigrateParam struct {
	NewAsset common.Address
}

func (this *MigrateParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.NewAsset); err != nil {
		return fmt.Errorf("serialize: write new asset addr failed, err: %s", err)
	}
	return nil
}

func (this *MigrateParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.NewAsset, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read new asset addr failed, err: %s", err)
	}
	return nil
}

type BalanceParam struct {
	User common.Address
}

func (this *BalanceParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user addr failed, err: %s", err)
	}
	return nil
}

func (this *BalanceParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	return nil
}

type AllowanceParam struct {
	Owner   common.Address
	Spender common.Address
}

func (this *AllowanceParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("serialize: write owner addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Spender); err != nil {
		return fmt.Errorf("serialize: write spender addr failed, err: %s", err)
	}
	return nil
}

func (this *AllowanceParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner addr failed, err: %s", err)
	}
	if this.Spender, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read spender addr failed, err: %s", err)
	}
	return nil
}

type MintParam struct {
	User   common.Address
	Amount *big.Int
}

func (this *MintParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *MintParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type BurnParam struct {
	User   common.Address
	Amount *big.Int
}

func (this *BurnParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *BurnParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type TransferParam struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

func (this *TransferParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("serialize: write from addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("serialize: write to addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *TransferParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from addr failed, err: %s", err)
	}
	if this.To, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read to addr failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type MultiTransferParam struct {
	Transfers []*TransferParam
}

func (this *MultiTransferParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.Transfers))); err != nil {
		return fmt.Errorf("serialize: write transfers num failed, err: %s", err)
	}
	for i, tran := range this.Transfers {
		if err := tran.Serialize(w); err != nil {
			return fmt.Errorf("serialize: write transfer failed, index %d, err: %s", i, err)
		}
	}
	return nil
}

func (this *MultiTransferParam) Deserialize(r io.Reader) error {
	var err error = nil
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read transfers num failed, err: %s", err)
	}
	this.Transfers = make([]*TransferParam, num)
	for i := uint64(0); i < num; i++ {
		tran := &TransferParam{}
		if err := tran.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize: read transfer failed, index %d, err: %s", i, err)
		}
		this.Transfers[i] = tran
	}
	return nil
}

type ApproveParam struct {
	Owner     common.Address
	Spender   common.Address
	Allowance *big.Int
}

func (this *ApproveParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("serialize: write owner addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Spender); err != nil {
		return fmt.Errorf("serialize: write spender addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Allowance)); err != nil {
		return fmt.Errorf("serialize: write allowance failed, err: %s", err)
	}
	return nil
}

func (this *ApproveParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner addr failed, err: %s", err)
	}
	if this.Spender, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read spender addr failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read allowance failed, err: %s", err)
	} else {
		this.Allowance = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type TransferFromParam struct {
	Spender common.Address
	From    common.Address
	To      common.Address
	Amount  *big.Int
}

func (this *TransferFromParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Spender); err != nil {
		return fmt.Errorf("serialize: write spender addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("serialize: write from addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("serialize: write to addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *TransferFromParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Spender, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read spender addr failed, err: %s", err)
	}
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from addr failed, err: %s", err)
	}
	if this.To, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read to addr failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type XShardTransferParam struct {
	From    common.Address
	To      common.Address
	ToShard common.ShardID
	Amount  *big.Int
}

func (this *XShardTransferParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("serialize: write from addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("serialize: write to addr failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ToShard); err != nil {
		return fmt.Errorf("serialize: write to shard id failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *XShardTransferParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from addr failed, err: %s", err)
	}
	if this.To, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read to addr failed, err: %s", err)
	}
	if this.ToShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type XShardTransferRetryParam struct {
	From       common.Address
	TransferId *big.Int
}

func (this *XShardTransferRetryParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("serialize: write from addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.TransferId)); err != nil {
		return fmt.Errorf("serialize: write transfer id failed, err: %s", err)
	}
	return nil
}

func (this *XShardTransferRetryParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.From, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from addr failed, err: %s", err)
	}
	if id, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read transfer id failed, err: %s", err)
	} else {
		this.TransferId = common.BigIntFromNeoBytes(id)
	}
	return nil
}

type ShardMintParam struct {
	Asset       uint64
	Account     common.Address
	FromShard   common.ShardID
	FromAccount common.Address
	TransferId  *big.Int
	Amount      *big.Int
}

func (this *ShardMintParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Asset); err != nil {
		return fmt.Errorf("serialize: write asset id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Account); err != nil {
		return fmt.Errorf("serialize: write account addr failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.FromShard); err != nil {
		return fmt.Errorf("serialize: write from shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.FromAccount); err != nil {
		return fmt.Errorf("serialize: write from account addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.TransferId)); err != nil {
		return fmt.Errorf("serialize: write transfer id failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.Amount)); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *ShardMintParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Asset, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read asset id failed, err: %s", err)
	}
	if this.Account, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read account addr failed, err: %s", err)
	}
	if this.FromShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read from shard failed, err: %s", err)
	}
	if this.FromAccount, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from account addr failed, err: %s", err)
	}
	if id, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read transfer id failed, err: %s", err)
	} else {
		this.TransferId = common.BigIntFromNeoBytes(id)
	}
	if amount, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	} else {
		this.Amount = common.BigIntFromNeoBytes(amount)
	}
	return nil
}

type XShardTranSuccParam struct {
	Asset      uint64
	Account    common.Address
	TransferId *big.Int
}

func (this *XShardTranSuccParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Asset); err != nil {
		return fmt.Errorf("serialize: write asset id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Account); err != nil {
		return fmt.Errorf("serialize: write account addr failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.TransferId)); err != nil {
		return fmt.Errorf("serialize: write transfer id failed, err: %s", err)
	}
	return nil
}

func (this *XShardTranSuccParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Asset, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read asset id failed, err: %s", err)
	}
	if this.Account, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read account addr failed, err: %s", err)
	}
	if id, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read transfer id failed, err: %s", err)
	} else {
		this.TransferId = common.BigIntFromNeoBytes(id)
	}
	return nil
}

type GetXShardTransferInfoParam struct {
	Account    common.Address
	Asset      uint64
	TransferId *big.Int
}

func (this *GetXShardTransferInfoParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Account); err != nil {
		return fmt.Errorf("serialize: write account addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Asset); err != nil {
		return fmt.Errorf("serialize: write asset id failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(this.TransferId)); err != nil {
		return fmt.Errorf("serialize: write transfer id failed, err: %s", err)
	}
	return nil
}

func (this *GetXShardTransferInfoParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Account, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read account addr failed, err: %s", err)
	}
	if this.Asset, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read asset id failed, err: %s", err)
	}
	if id, err := serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read transfer id failed, err: %s", err)
	} else {
		this.TransferId = common.BigIntFromNeoBytes(id)
	}
	return nil
}

type GetPendingXShardTransferParam struct {
	Account common.Address
	Asset   uint64
}

func (this *GetPendingXShardTransferParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Account); err != nil {
		return fmt.Errorf("serialize: write account addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Asset); err != nil {
		return fmt.Errorf("serialize: write asset id failed, err: %s", err)
	}
	return nil
}

func (this *GetPendingXShardTransferParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Account, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read account addr failed, err: %s", err)
	}
	if this.Asset, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read asset id failed, err: %s", err)
	}
	return nil
}
