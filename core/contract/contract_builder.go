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
	"math/big"
	"sort"

	"github.com/Ontology/common"
	pg "github.com/Ontology/core/contract/program"
	"github.com/Ontology/errors"
	vm "github.com/Ontology/vm/neovm"
	"github.com/ontio/ontology-crypto/keypair"
)

//create a Single Singature contract for owner
func CreateSignatureContract(ownerPubKey keypair.PublicKey) (*Contract, error) {
	temp := keypair.SerializePublicKey(ownerPubKey)
	signatureRedeemScript, err := CreateSignatureRedeemScript(ownerPubKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash := common.ToCodeHash(temp)
	signatureRedeemScriptHashToCodeHash := common.ToCodeHash(signatureRedeemScript)
	return &Contract{
		Code:            signatureRedeemScript,
		Parameters:      []ContractParameterType{Signature},
		ProgramHash:     signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: hash,
	}, nil
}

func CreateSignatureRedeemScript(pubkey keypair.PublicKey) ([]byte, error) {
	temp := keypair.SerializePublicKey(pubkey)
	sb := pg.NewProgramBuilder()
	sb.PushData(temp)
	sb.AddOp(vm.CHECKSIG)
	return sb.ToArray(), nil
}

//create a Multi Singature contract for owner  。
func CreateMultiSigContract(publicKeyHash common.Address, m int, publicKeys []keypair.PublicKey) (*Contract, error) {

	params := make([]ContractParameterType, m)
	for i, _ := range params {
		params[i] = Signature
	}
	MultiSigRedeemScript, err := CreateMultiSigRedeemScript(m, publicKeys)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[Contract],CreateSignatureRedeemScript failed.")
	}
	signatureRedeemScriptHashToCodeHash := common.ToCodeHash(MultiSigRedeemScript)
	return &Contract{
		Code:            MultiSigRedeemScript,
		Parameters:      params,
		ProgramHash:     signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: publicKeyHash,
	}, nil
}

func CreateMultiSigRedeemScript(m int, pubkeys []keypair.PublicKey) ([]byte, error) {
	if !(m >= 1 && m <= len(pubkeys) && len(pubkeys) <= 24) {
		return nil, nil //TODO: add panic
	}

	sb := pg.NewProgramBuilder()
	sb.PushNumber(big.NewInt(int64(m)))

	//sort pubkey
	sort.Sort(keypair.NewPublicList(pubkeys))

	for _, pubkey := range pubkeys {
		temp := keypair.SerializePublicKey(pubkey)
		sb.PushData(temp)
	}

	sb.PushNumber(big.NewInt(int64(len(pubkeys))))
	sb.AddOp(vm.CHECKMULTISIG)
	return sb.ToArray(), nil
}
