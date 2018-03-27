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

package contract

import (
	"errors"
	"math/big"
	"sort"

	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	pg "github.com/Ontology/core/contract/program"
	sig "github.com/Ontology/core/signature"
	_ "github.com/Ontology/errors"
	"github.com/ontio/ontology-crypto/keypair"
)

type ContractContext struct {
	Data          sig.SignableData
	ProgramHashes []Address
	Codes         [][]byte
	Parameters    [][][]byte

	MultiPubkeyPara [][]PubkeyParameter

	//temp index for multi sig
	tempParaIndex int
}

func NewContractContext(data sig.SignableData) *ContractContext {
	programHashes, _ := data.GetProgramHashes() //TODO: check error
	log.Debug("programHashes= ", programHashes)
	log.Debug("hashLen := len(programHashes) ", len(programHashes))
	hashLen := len(programHashes)
	return &ContractContext{
		Data:            data,
		ProgramHashes:   programHashes,
		Codes:           make([][]byte, hashLen),
		Parameters:      make([][][]byte, hashLen),
		MultiPubkeyPara: make([][]PubkeyParameter, hashLen),
		tempParaIndex:   0,
	}
}

func (cxt *ContractContext) Add(contract *Contract, index int, parameter []byte) error {
	log.Debug()
	i := cxt.GetIndex(contract.ProgramHash)
	if i < 0 {
		log.Warn("Program Hash is not exist, using 0 by default")
		i = 0
	}
	if cxt.Codes[i] == nil {
		cxt.Codes[i] = contract.Code
	}
	if cxt.Parameters[i] == nil {
		cxt.Parameters[i] = make([][]byte, len(contract.Parameters))
	}
	cxt.Parameters[i][index] = parameter
	return nil
}

func (cxt *ContractContext) AddContract(contract *Contract, pubkey keypair.PublicKey, parameter []byte) error {
	log.Debug()
	if contract.GetType() == MultiSigContract {
		log.Debug()
		// add multi sig contract

		log.Debug("Multi Sig: contract.ProgramHash:", contract.ProgramHash)
		log.Debug("Multi Sig: cxt.ProgramHashes:", cxt.ProgramHashes)

		index := cxt.GetIndex(contract.ProgramHash)

		log.Debug("Multi Sig: GetIndex:", index)

		if index < 0 {
			log.Error("The program hash is not exist.")
			return errors.New("The program hash is not exist.")
		}

		log.Debug("Multi Sig: contract.Code:", cxt.Codes[index])

		if cxt.Codes[index] == nil {
			cxt.Codes[index] = contract.Code
		}
		log.Debug("Multi Sig: cxt.Codes[index]:", cxt.Codes[index])

		if cxt.Parameters[index] == nil {
			cxt.Parameters[index] = make([][]byte, len(contract.Parameters))
		}
		log.Debug("Multi Sig: cxt.Parameters[index]:", cxt.Parameters[index])

		if err := cxt.Add(contract, cxt.tempParaIndex, parameter); err != nil {
			return err
		}

		cxt.tempParaIndex++

		//all paramenters added, sort the parameters
		if cxt.tempParaIndex == len(contract.Parameters) {
			cxt.tempParaIndex = 0
		}

		//TODO: Sort the parameter according contract's PK list sequence
		//if err := cxt.AddSignatureToMultiList(index,contract,pubkey,parameter); err != nil {
		//	return err
		//}
		//
		//if(cxt.tempParaIndex == len(contract.Parameters)){
		//	//all multi sigs added, sort the sigs and add to context
		//	if err := cxt.AddMultiSignatures(index,contract,pubkey,parameter);err != nil {
		//		return err
		//	}
		//}

	} else {
		//add non multi sig contract
		log.Debug()
		index := -1
		for i := 0; i < len(contract.Parameters); i++ {
			if contract.Parameters[i] == Signature {
				if index >= 0 {
					return errors.New("Contract Parameters are not supported.")
				} else {
					index = i
				}
			}
		}
		return cxt.Add(contract, index, parameter)
	}
	return nil
}

func (cxt *ContractContext) AddSignatureToMultiList(contractIndex int, contract *Contract, pubkey keypair.PublicKey, parameter []byte) error {
	if cxt.MultiPubkeyPara[contractIndex] == nil {
		cxt.MultiPubkeyPara[contractIndex] = make([]PubkeyParameter, len(contract.Parameters))
	}
	pk := keypair.SerializePublicKey(pubkey)

	pubkeyPara := PubkeyParameter{
		PubKey:    ToHexString(pk),
		Parameter: ToHexString(parameter),
	}
	cxt.MultiPubkeyPara[contractIndex] = append(cxt.MultiPubkeyPara[contractIndex], pubkeyPara)

	return nil
}

func (cxt *ContractContext) AddMultiSignatures(index int, contract *Contract, pubkey keypair.PublicKey, parameter []byte) error {
	pkIndexs, err := cxt.ParseContractPubKeys(contract)
	if err != nil {
		return errors.New("Contract Parameters are not supported.")
	}

	paraIndexs := []ParameterIndex{}
	for _, pubkeyPara := range cxt.MultiPubkeyPara[index] {
		pubKeyBytes, err := HexToBytes(pubkeyPara.Parameter)
		if err != nil {
			return errors.New("Contract AddContract pubKeyBytes HexToBytes failed.")
		}

		paraIndex := ParameterIndex{
			Parameter: pubKeyBytes,
			Index:     pkIndexs[pubkeyPara.PubKey],
		}
		paraIndexs = append(paraIndexs, paraIndex)
	}

	//sort parameter by Index
	sort.Sort(sort.Reverse(ParameterIndexSlice(paraIndexs)))

	//generate sorted parameter list
	for i, paraIndex := range paraIndexs {
		if err := cxt.Add(contract, i, paraIndex.Parameter); err != nil {
			return err
		}
	}

	cxt.MultiPubkeyPara[index] = nil

	return nil
}

func (cxt *ContractContext) ParseContractPubKeys(contract *Contract) (map[string]int, error) {

	pubkeyIndex := make(map[string]int)

	Index := 0
	//parse contract's pubkeys
	i := 0
	switch contract.Code[i] {
	case 1:
		i += 2
		break
	case 2:
		i += 3
		break
	}
	for contract.Code[i] == 33 {
		i++

		//add to parameter index
		pubkeyIndex[ToHexString(contract.Code[i:33])] = Index

		i += 33
		Index++
	}

	return pubkeyIndex, nil
}

func (cxt *ContractContext) GetIndex(programHash Address) int {
	for i := 0; i < len(cxt.ProgramHashes); i++ {
		if cxt.ProgramHashes[i] == programHash {
			return i
		}
	}
	return -1
}

func (cxt *ContractContext) GetPrograms() []*pg.Program {
	log.Debug()
	//log.Debug("!cxt.IsCompleted()=",!cxt.IsCompleted())
	//log.Debug(cxt.Codes)
	//log.Debug(cxt.Parameters)
	if !cxt.IsCompleted() {
		return nil
	}
	programs := make([]*pg.Program, len(cxt.Parameters))

	log.Debug(" len(cxt.Codes)", len(cxt.Codes))

	for i := 0; i < len(cxt.Codes); i++ {
		sb := pg.NewProgramBuilder()

		for _, parameter := range cxt.Parameters[i] {
			if len(parameter) <= 2 {
				sb.PushNumber(new(big.Int).SetBytes(parameter))
			} else {
				sb.PushData(parameter)
			}
		}
		//log.Debug(" cxt.Codes[i])", cxt.Codes[i])
		//log.Debug(" sb.ToArray()", sb.ToArray())
		programs[i] = &pg.Program{
			Code:      cxt.Codes[i],
			Parameter: sb.ToArray(),
		}
	}
	return programs
}

func (cxt *ContractContext) IsCompleted() bool {
	for _, p := range cxt.Parameters {
		if p == nil {
			return false
		}

		for _, pp := range p {
			if pp == nil {
				return false
			}
		}
	}
	return true
}
