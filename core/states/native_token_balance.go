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

func (self NativeTokenBalance) MustToStorageItemBytes() []byte {
	return self.MustToStorageItem().ToArray()
}

func (self NativeTokenBalance) Add(rhs NativeTokenBalance) NativeTokenBalance {
	return NativeTokenBalance{Balance: self.Balance.Add(rhs.Balance)}
}

func (self NativeTokenBalance) MustSub(rhs NativeTokenBalance) NativeTokenBalance {
	if self.Balance.LessThan(rhs.Balance) {
		panic(fmt.Errorf("balance sub underflow: a: %s, b: %s", self.Balance, rhs.Balance))
	}

	return NativeTokenBalance{Balance: self.Balance.Sub(rhs.Balance)}
}

func (self NativeTokenBalance) MustToStorageItem() *StorageItem {
	if self.IsFloat() {
		return &StorageItem{
			StateBase: StateBase{StateVersion: ScaleDecimal9Version},
			Value:     self.Balance.BigInt().Bytes(),
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

	return NativeTokenBalance{Balance: bigint.New(big.NewInt(0).SetBytes(val.Value))}, nil
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
