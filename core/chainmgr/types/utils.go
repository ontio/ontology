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
package types

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/states"
	sComm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
)

func getRawStorageItemFromMemDb(memdb *overlaydb.MemDB, addr common.Address, key []byte) (value []byte, unkown bool) {
	rawKey := make([]byte, 0, 1+common.ADDR_LEN+len(key))
	rawKey = append(rawKey, byte(sComm.ST_STORAGE))
	rawKey = append(rawKey, addr[:]...)
	rawKey = append(rawKey, key...)
	return memdb.Get(rawKey)
}

func GetStorageValue(memdb *overlaydb.MemDB, backend *ledger.Ledger, addr common.Address, key []byte) (value []byte, err error) {
	if memdb == nil {
		return backend.GetStorageItem(addr, key)
	}
	rawValue, unknown := getRawStorageItemFromMemDb(memdb, addr, key)
	if unknown {
		return backend.GetStorageItem(addr, key)
	}
	if len(rawValue) == 0 {
		return nil, sComm.ErrNotFound
	}

	value, err = states.GetValueFromRawStorageItem(rawValue)
	return
}
