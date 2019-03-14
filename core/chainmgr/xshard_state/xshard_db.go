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

package xshard_state

import (
	"encoding/hex"

	"github.com/ontio/ontology/core/store/common"
)

//
// xShardMsgKV: kv store of shard-message-queue contract
//

var xShardMsgKV = make(map[string][]byte)

func GetKVStorageItem(key []byte) ([]byte, error) {
	k := hex.EncodeToString(key)
	if v, present := xShardMsgKV[k]; !present {
		return nil, common.ErrNotFound
	} else {
		return v, nil
	}
}

func PutKV(key, value []byte) {
	k := hex.EncodeToString(key)
	xShardMsgKV[k] = value
}
