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

package types

import (
	"crypto/sha256"
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/program"
	"golang.org/x/crypto/ripemd160"
)

func AddressFromPubKey(pubkey keypair.PublicKey) common.Address {
	prog := program.ProgramFromPubKey(pubkey)

	return AddressFromVmCode(prog)
}

func AddressFromMultiPubKeys(pubkeys []keypair.PublicKey, m int) (common.Address, error) {
	var addr common.Address
	n := len(pubkeys)
	if !(1 <= m && m <= n && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		return addr, errors.New("wrong multi-sig param")
	}

	prog, err := program.ProgramFromMultiPubKey(pubkeys, m)
	if err != nil {
		return addr, err
	}

	return AddressFromVmCode(prog), nil
}

func AddressFromVmCode(code []byte) common.Address {
	var addr common.Address
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])

	return addr
}

func AddressFromBookkeepers(bookkeepers []keypair.PublicKey) (common.Address, error) {
	return AddressFromMultiPubKeys(bookkeepers, len(bookkeepers)-(len(bookkeepers)-1)/3)
}
