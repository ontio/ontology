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

package services
/*

import (
	//"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/common"
	"fmt"
	"errors"
	"github.com/Ontology/vm/wasmvm/memory"
)

func getBlockHash(param []uint64)uint64{
	return uint64(0)
}

func getBlockHight(param []uint64)uint64{
	return uint64(0)
}

func getValue(param []uint64)uint64{
	return uint64(0)
}

func storeValue(param []uint64)uint64{
	return uint64(0)
}

func  blockChainGetHeight(mem *memory.VMmemory,param []uint64)(uint64,error){
	var i uint32
	if ledger.DefaultLedger == nil {
		i = 0
	} else {
		i = ledger.DefaultLedger.Store.GetHeight()
	}
	return uint64(i),nil
}


func  getAddress(memory *memory.VMmemory,param []uint64)(uint64,error){
	mem := memory.Memory
	addr:=mem[:32]
	u256,_:=common.Uint256ParseFromBytes(addr)
	fmt.Println("getAdderss u256:",u256.ToString())

	return uint64(356),nil
}


func blockChainGetBlockHash(memory *memory.VMmemory,param []uint64)(uint64,error){

	mem := memory.Memory
	if len(param) == 2{ //block height
		height := uint32(param[0])
		offset := int(uint32(param[1]))
		if ledger.DefaultLedger != nil {
			hash, err := ledger.DefaultLedger.Store.GetBlockHash(height)
			if err != nil {
				return uint64(0), errors.New("[BlockChainGetHeader] GetBlockHash error!.")
			}
			hashbytes := hash.ToArray()
			copy(mem[offset+4:offset+4+len(hashbytes)],hashbytes)
			return uint64(len(hashbytes)),nil

		}else{
			return uint64(0),errors.New("get default ledger failed")
		}
	}else{
		return uint64(0),errors.New("[BlockChainGetHeader] GetBlockHash parameters error!.")
	}

	//only for test
	//s:="1234567890"
	//offset := int(uint32(param[1]))
	//b:= []byte(s)
	//copy(mem[offset:offset+10],b)
	//return uint64(10),nil
}
*/
