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

package cross_chain

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/vm/neovm/types"
)

const (
	CREATE_CROSS_CHAIN_TX  = "createCrossChainTx"
	PROCESS_CROSS_CHAIN_TX = "processCrossChainTx"
	MAKE_FROM_ONT_PROOF    = "makeFromOntProof"
	VERIFY_TO_ONT_PROOF    = "verifyToOntProof"

	//key prefix
	DONE_TX = "doneTx"
	REQUEST = "request"

	//ont chain id
	ONT_CHAIN_ID = 2
)

//Init governance contract address
func InitCrossChain() {
	native.Contracts[utils.CrossChainContractAddress] = RegisterCrossChianContract
}

//Register methods of governance contract
func RegisterCrossChianContract(native *native.NativeService) {
	native.Register(CREATE_CROSS_CHAIN_TX, CreateCrossChainTx)
	native.Register(PROCESS_CROSS_CHAIN_TX, ProcessCrossChainTx)
}

func CreateCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(CreateCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}

	err := MakeFromOntProof(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, MakeOntProof error: %v", err)
	}

	//TODO: miner fee?

	return utils.BYTE_TRUE, nil
}

func ProcessCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(ProcessCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, contract params deserialize error: %v", err)
	}

	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, proof hex.DecodeString error: %v", err)
	}
	merkleValue, err := VerifyToOntTx(native, proof, params.FromChainID, params.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, VerifyOntTx error: %v", err)
	}

	if merkleValue.MakeTxParam.ToChainID != ONT_CHAIN_ID {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, to chain id is not ont")
	}

	//call cross chain function
	dest, err := common.AddressParseFromBytes(merkleValue.MakeTxParam.ToContractAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressParseFromBytes error: %v", err)
	}
	functionName := merkleValue.MakeTxParam.Method
	fromContractAddress := merkleValue.MakeTxParam.FromContractAddress
	args := merkleValue.MakeTxParam.Args

	res, err := native.NeoVMCall(dest, functionName, args, fromContractAddress, merkleValue.FromChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NeoVMCall error: %v", err)
	}
	r, ok := res.(*types.Integer)
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, res of neo vm call must be bool")
	}
	v, _ := r.GetBigInteger()
	if v.Cmp(new(big.Int).SetUint64(0)) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, res of neo vm call is false")
	}

	//TODO: miner fee?

	return utils.BYTE_TRUE, nil
}
