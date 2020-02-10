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
package wasmvm

import (
	"errors"
	"math"

	"github.com/ontio/ontology/core/states"
	"github.com/ontio/wagon/exec"
)

func storageRead(service *WasmVmService, keybytes []byte, klen uint32, vlen uint32, offset uint32) ([]byte, uint32, error) {
	key := serializeStorageKey(service.ContextRef.CurrentContext().ContractAddress, keybytes)

	raw, err := service.CacheDB.Get(key)
	if err != nil {
		return []byte{}, 0, err
	}

	if raw == nil {
		return []byte{}, math.MaxUint32, nil
	}

	item, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return []byte{}, 0, err
	}

	length := vlen
	itemlen := uint32(len(item))
	if itemlen < vlen {
		length = itemlen
	}

	if uint32(len(item)) < offset {
		return []byte{}, 0, errors.New("offset is invalid")
	}

	return item[offset : offset+length], uint32(len(item)), nil
}

func StorageRead(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32 {
	self := proc.HostData().(*Runtime)
	self.checkGas(STORAGE_GET_GAS)
	keybytes, err := ReadWasmMemory(proc, keyPtr, klen)
	if err != nil {
		panic(err)
	}

	itemWrite, originLen, err := storageRead(self.Service, keybytes, klen, vlen, offset)
	if err != nil {
		panic(err)
	}

	if originLen != math.MaxUint32 {
		_, err = proc.WriteAt(itemWrite[:], int64(val))

		if err != nil {
			panic(err)
		}
	}

	return originLen
}

func StorageWrite(proc *exec.Process, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) {
	self := proc.HostData().(*Runtime)
	keybytes, err := ReadWasmMemory(proc, keyPtr, keyLen)
	if err != nil {
		panic(err)
	}

	valbytes, err := ReadWasmMemory(proc, valPtr, valLen)
	if err != nil {
		panic(err)
	}

	cost := uint64((len(keybytes)+len(valbytes)-1)/1024+1) * STORAGE_PUT_GAS
	self.checkGas(cost)

	key := serializeStorageKey(self.Service.ContextRef.CurrentContext().ContractAddress, keybytes)

	self.Service.CacheDB.Put(key, states.GenRawStorageItem(valbytes))
}

func StorageDelete(proc *exec.Process, keyPtr uint32, keyLen uint32) {
	self := proc.HostData().(*Runtime)
	self.checkGas(STORAGE_DELETE_GAS)
	keybytes, err := ReadWasmMemory(proc, keyPtr, keyLen)
	if err != nil {
		panic(err)
	}
	key := serializeStorageKey(self.Service.ContextRef.CurrentContext().ContractAddress, keybytes)

	self.Service.CacheDB.Delete(key)
}
