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

package event

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestDeserialization(t *testing.T) {
	data, err := hex.DecodeString("000000000000000000000000000000000000000203000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef000000000000000000000000a34886547e00d8f15eaf5a98f99f4f76aaeb3bd500000000000000000000000000000000000000000000000000000000000000072000000000000000000000000000000000000000000000000000ffa4f70a6cd800")
	if err != nil {
		panic(err)
	}
	source := common.NewZeroCopySource(data)
	sl := &types.StorageLog{}
	sl.Deserialization(source)
	fmt.Println(sl.Address.String())
	a := big.NewInt(0).SetBytes(sl.Data)
	fmt.Println(a.String())
	info := NotifyEventInfoFromEvmLog(sl)
	sl2, err := NotifyEventInfoToEvmLog(info)
	assert.Nil(t, err)
	fmt.Println(sl2)
}
