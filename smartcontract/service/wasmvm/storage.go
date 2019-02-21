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
	"github.com/go-interpreter/wagon/exec"
	"github.com/pkg/errors"
	"math"
)

func (self *Runtime) StorageRead(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32 {
	self.checkGas(STORAGE_GET_GAS)
	keybytes := make([]byte, klen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}

	key, err := serializeStorageKey(self.Service.ContextRef.CurrentContext().ContractAddress, keybytes)
	if err != nil {
		panic(err)
	}
	item, err := self.Service.CacheDB.Get(key)
	if err != nil {
		panic(err)
	}

	if item == nil {
		return math.MaxUint32
	}

	length, err := proc.WriteAt(item, int64(val))
	if err != nil {
		panic(err)
	}
	return uint32(length)
}

func (self *Runtime) StorageWrite(proc *exec.Process, keyPtr uint32, keylen uint32, valPtr uint32, valLen uint32) {

	keybytes := make([]byte, keylen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}
	if len(keybytes) > 1024 {
		panic(errors.New("[storageWrite]:key should not longer than 1024 bytes"))
	}

	valbytes := make([]byte, valLen)
	_, err = proc.ReadAt(valbytes, int64(valPtr))
	if err != nil {
		panic(err)
	}

	cost := uint64(((len(keybytes)+len(valbytes)-1)/1024 + 1)) * STORAGE_PUT_GAS
	self.checkGas(cost)

	key, err := serializeStorageKey(self.Service.ContextRef.CurrentContext().ContractAddress, keybytes)
	if err != nil {
		panic(err)
	}

	self.Service.CacheDB.Put(key, valbytes)
}

func (self *Runtime) StorageDelete(proc *exec.Process, keyPtr uint32, keylen uint32) {
	self.checkGas(STORAGE_DELETE_GAS)
	keybytes := make([]byte, keylen)
	_, err := proc.ReadAt(keybytes, int64(keyPtr))
	if err != nil {
		panic(err)
	}
	key, err := serializeStorageKey(self.Service.ContextRef.CurrentContext().ContractAddress, keybytes)
	if err != nil {
		panic(err)
	}
	self.Service.CacheDB.Delete(key)
}
