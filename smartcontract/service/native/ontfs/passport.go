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

package ontfs

import (
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type Passport struct {
	BlockHeight uint64
	BlockHash   []byte
	WalletAddr  common.Address
	PublicKey   []byte
	Signature   []byte
}

func (this *Passport) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.BlockHeight)
	sink.WriteVarBytes(this.BlockHash)
	utils.EncodeAddress(sink, this.WalletAddr)
	sink.WriteVarBytes(this.PublicKey)
	sink.WriteVarBytes(this.Signature)
}

func (this *Passport) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.BlockHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.BlockHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.WalletAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PublicKey, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.Signature, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	return nil
}

func CheckPassport(currBlockHeight uint64, passportData []byte) (common.Address, error) {
	var err error
	var passport Passport
	src := common.NewZeroCopySource(passportData)
	if err = passport.Deserialization(src); err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("CheckPassport Deserialization error")
	}

	if passport.BlockHeight > currBlockHeight || passport.BlockHeight+DefaultPassportExpire < currBlockHeight {
		return passport.WalletAddr, fmt.Errorf("CheckPassport passport expired")
	}

	pubKey, err := keypair.DeserializePublicKey(passport.PublicKey)
	if err != nil {
		return passport.WalletAddr, fmt.Errorf("CheckPassport DeserializePublicKey error: %s", err.Error())
	}

	addr := types.AddressFromPubKey(pubKey)
	if addr != passport.WalletAddr {
		return passport.WalletAddr, fmt.Errorf("CheckPassport Pubkey not match walletAddr ")
	}

	passportTmp := Passport{
		BlockHeight: passport.BlockHeight,
		BlockHash:   passport.BlockHash,
		WalletAddr:  passport.WalletAddr,
		PublicKey:   passport.PublicKey,
	}

	sink := common.NewZeroCopySink(nil)
	passportTmp.Serialization(sink)

	signValue, err := signature.Deserialize(passport.Signature)
	if err != nil {
		return passport.WalletAddr, fmt.Errorf("CheckPassport signature Deserialize error: %s", err.Error())
	}

	if signature.Verify(pubKey, sink.Bytes(), signValue) {
		return passport.WalletAddr, nil
	} else {
		return passport.WalletAddr, fmt.Errorf("CheckPassport Verify error: %s", err.Error())
	}
}
