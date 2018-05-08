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
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

/* crypto object */
type Account struct {
	Label      string
	KeyType    string
	Curve      string
	IsDefault  bool
	PrivateKey keypair.PrivateKey
	PublicKey  keypair.PublicKey
	Address    common.Address
	SigScheme  s.SignatureScheme
}

func NewAccount(encrypt string) *Account {
	// Determine the public key algorithm and parameters according to
	// the encrypt.
	var pkAlgorithm keypair.KeyType
	var params interface{}
	var scheme s.SignatureScheme
	var err error
	if "" != encrypt {
		scheme, err = s.GetScheme(encrypt)
	} else {
		scheme = s.SHA256withECDSA
	}
	if err != nil {
		log.Warn("unknown signature scheme, use SHA256withECDSA as default.")
		scheme = s.SHA256withECDSA
	}
	keyType := ""
	curve := ""
	switch scheme {
	case s.SHA224withECDSA, s.SHA3_224withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P224
		keyType = "ecdsa"
		curve = "P-224"
	case s.SHA256withECDSA, s.SHA3_256withECDSA, s.RIPEMD160withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P256
		keyType = "ecdsa"
		curve = "P-256"
	case s.SHA384withECDSA, s.SHA3_384withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P384
		keyType = "ecdsa"
		curve = "P-384"
	case s.SHA512withECDSA, s.SHA3_512withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P521
		keyType = "ecdsa"
		curve = "P-521"
	case s.SM3withSM2:
		pkAlgorithm = keypair.PK_SM2
		params = keypair.SM2P256V1
		keyType = "sm2"
		curve = "SM2P256V1"
	case s.SHA512withEDDSA:
		pkAlgorithm = keypair.PK_EDDSA
		params = keypair.ED25519
		keyType = "ed25519"
		curve = "ED25519"
	}

	pri, pub, _ := keypair.GenerateKeyPair(pkAlgorithm, params)
	address := types.AddressFromPubKey(pub)
	return &Account{
		KeyType:    keyType,
		Curve:      curve,
		PrivateKey: pri,
		PublicKey:  pub,
		Address:    address,
		SigScheme:  scheme,
	}
}

func (this *Account) PrivKey() keypair.PrivateKey {
	return this.PrivateKey
}

func (this *Account) PubKey() keypair.PublicKey {
	return this.PublicKey
}

func (this *Account) Scheme() s.SignatureScheme {
	return this.SigScheme
}

//AccountPublic without private key
type AccountPublic struct {
	Label     string
	KeyType   string
	Curve     string
	CipherKey []byte
	IsDefault bool
	PublicKey keypair.PublicKey
	Address   common.Address
	SigScheme s.SignatureScheme
}
