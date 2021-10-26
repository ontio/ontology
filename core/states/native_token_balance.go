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

package states

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/laizy/bigint"
	"github.com/ontio/ontology/common"
)

const ScaleFactor = 1000000000
const ScaleDecimal9Version = 1
const DefaultVersion = 0

// ont balance with decimal 9
type NativeTokenBalance struct {
	Balance bigint.Int
}

func (self *NativeTokenBalance) String() string {
	return self.Balance.String()
}

func (self NativeTokenBalance) MustToStorageItemBytes() []byte {
	return self.MustToStorageItem().ToArray()
}

func (self NativeTokenBalance) Add(rhs NativeTokenBalance) NativeTokenBalance {
	return NativeTokenBalance{Balance: self.Balance.Add(rhs.Balance)}
}

func (self NativeTokenBalance) IsZero() bool {
	return self.Balance.IsZero()
}

func (self NativeTokenBalance) Sub(rhs NativeTokenBalance) (result NativeTokenBalance, err error) {
	if self.Balance.LessThan(rhs.Balance) {
		return result, fmt.Errorf("balance sub underflow: a: %s, b: %s", self.Balance, rhs.Balance)
	}
	return NativeTokenBalance{Balance: self.Balance.Sub(rhs.Balance)}, nil
}

func (self NativeTokenBalance) MustToStorageItem() *StorageItem {
	if self.IsFloat() {
		return &StorageItem{
			StateBase: StateBase{StateVersion: ScaleDecimal9Version},
			Value:     common.BigIntToNeoBytes(self.Balance.BigInt()),
		}
	}

	return &StorageItem{
		StateBase: StateBase{StateVersion: DefaultVersion},
		Value:     common.NewZeroCopySink(nil).WriteUint64(self.MustToInteger64()).Bytes(),
	}
}

func NativeTokenBalanceFromStorageItem(val *StorageItem) (NativeTokenBalance, error) {
	if val.StateVersion == DefaultVersion {
		balance, err := common.NewZeroCopySource(val.Value).ReadUint64()
		if err != nil {
			return NativeTokenBalance{}, err
		}
		return NativeTokenBalance{Balance: bigint.Mul(balance, ScaleFactor)}, nil
	}
	balance := bigint.New(common.BigIntFromNeoBytes(val.Value))
	if balance.IsNegative() {
		return NativeTokenBalance{}, errors.New("negative balance")
	}

	return NativeTokenBalance{Balance: bigint.New(common.BigIntFromNeoBytes(val.Value))}, nil
}

func (self NativeTokenBalance) IsFloat() bool {
	return !self.Balance.Mod(ScaleFactor).IsZero()
}

func (self NativeTokenBalance) ToInteger() bigint.Int {
	return self.Balance.Div(ScaleFactor)
}

func (self NativeTokenBalance) FloatPart() uint64 {
	val := self.Balance.Mod(ScaleFactor).BigInt()
	return val.Uint64()
}

func (self NativeTokenBalance) MustToInteger64() uint64 {
	val := self.Balance.Div(ScaleFactor).BigInt()
	if !val.IsUint64() {
		panic("too large token balance")
	}

	return val.Uint64()
}

func NativeTokenBalanceFromInteger(val uint64) NativeTokenBalance {
	return NativeTokenBalance{Balance: bigint.Mul(val, ScaleFactor)}
}

func (self NativeTokenBalance) ToBigInt() *big.Int {
	return self.Balance.BigInt()
}
