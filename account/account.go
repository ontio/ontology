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

package account

import (
	"errors"
	. "github.com/Ontology/common"
	"github.com/Ontology/crypto"
	"github.com/Ontology/core/types"
)

type Account struct {
	PrivateKey []byte
	PublicKey  *crypto.PubKey
	Address    Address
}

func NewAccount() *Account {
	priKey, pubKey, _ := crypto.GenKeyPair()
	address := types.AddressFromPubKey(&pubKey)
	return &Account{
		PrivateKey: priKey,
		PublicKey:  &pubKey,
		Address:    address,
	}
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error) {
	privKeyLen := len(privateKey)

	if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
		return nil, errors.New("Invalid private Key.")
	}

	pubKey := crypto.NewPubKey(privateKey)
	address := types.AddressFromPubKey(pubKey)

	return &Account{
		PrivateKey: privateKey,
		PublicKey:  pubKey,
		Address:    address,
	}, nil
}

func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
