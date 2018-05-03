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
	"crypto/sha256"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

/* crypto object */
type Account struct {
	PrivateKey keypair.PrivateKey
	PublicKey  keypair.PublicKey
	Address    common.Address
	SigScheme  s.SignatureScheme
}

func NewAccount(encrypt string) *Account {
	// Determine the public key algorithm and parameters according to
	// the config file.
	// FIXME: better to decouple from config file by inputing as arguments.
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
	switch scheme {
	case s.SHA224withECDSA, s.SHA3_224withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P224
	case s.SHA256withECDSA, s.SHA3_256withECDSA, s.RIPEMD160withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P256
	case s.SHA384withECDSA, s.SHA3_384withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P384
	case s.SHA512withECDSA, s.SHA3_512withECDSA:
		pkAlgorithm = keypair.PK_ECDSA
		params = keypair.P521
	case s.SM3withSM2:
		pkAlgorithm = keypair.PK_SM2
		params = keypair.SM2P256V1
	case s.SHA512withEDDSA:
		pkAlgorithm = keypair.PK_EDDSA
		params = keypair.ED25519
	}

	pri, pub, _ := keypair.GenerateKeyPair(pkAlgorithm, params)
	address := types.AddressFromPubKey(pub)
	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Address:    address,
		SigScheme:  scheme,
	}
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error) {
	pri, err := keypair.DeserializePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	pub := pri.Public()
	address := types.AddressFromPubKey(pub)
	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Address:    address,
	}, nil
}

func (ac *Account) PrivKey() keypair.PrivateKey {
	return ac.PrivateKey
}

func (ac *Account) PubKey() keypair.PublicKey {
	return ac.PublicKey
}

func (ac *Account) Scheme() s.SignatureScheme {
	return ac.SigScheme
}

/** Accountx - for wallet read and save, no crypto object included **/
type Accountx struct {
	keypair.ProtectedKey

	Label     string `json:"label"`
	PubKey    string `json:"publicKey"`
	SigSch    string `json:"signatureScheme"`
	IsDefault bool   `json:"isDefault"`
	Lock      bool   `json:"lock"`
	PassHash  string `json:"passwordHash"`
}

func (this *Accountx) SetKeyPair(keyinfo *keypair.ProtectedKey) {
	this.Address = keyinfo.Address
	this.EncAlg = keyinfo.EncAlg
	this.Alg = keyinfo.Alg
	this.Hash = keyinfo.Hash
	this.Key = keyinfo.Key
	this.Param = keyinfo.Param
}
func (this *Accountx) GetKeyPair() *keypair.ProtectedKey {
	var keyinfo = new(keypair.ProtectedKey)
	keyinfo.Address = this.Address
	keyinfo.EncAlg = this.EncAlg
	keyinfo.Alg = this.Alg
	keyinfo.Hash = this.Hash
	keyinfo.Key = this.Key
	keyinfo.Param = this.Param
	return keyinfo
}
func (this *Accountx) VerifyPassword(pwd []byte) bool {
	passwordHash := sha256.Sum256(pwd)
	if this.PassHash != hex.EncodeToString(passwordHash[:]) {
		return false
	}
	return true
}

func CreateAccount(optionType *string, optionCurve *string, optionScheme *string, password *[]byte) *Accountx {

	inputKeyTypeInfo := KeyTypeMap[*optionType]
	inputCurveInfo := CurveMap[*optionCurve]
	inputSchemeInfo := SchemeMap[*optionScheme]

	prvkey, pubkey, _ := keypair.GenerateKeyPair(inputKeyTypeInfo.Code, inputCurveInfo.Code)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()

	prvSecret, _ := keypair.EncryptPrivateKey(prvkey, address, *password)
	h := sha256.Sum256(*password)

	var acc = new(Accountx)
	acc.SetKeyPair(prvSecret)
	acc.SigSch = inputSchemeInfo.Name
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	acc.PassHash = hex.EncodeToString(h[:])

	return acc
}

//map info, to get some information easily
//todo: move to crypto package
type KeyTypeInfo struct {
	Name string
	Code keypair.KeyType
}

var KeyTypeMap = map[string]KeyTypeInfo{
	"":  {"ecdsa", keypair.PK_ECDSA},
	"1": {"ecdsa", keypair.PK_ECDSA},
	"2": {"sm2", keypair.PK_SM2},
	"3": {"ed25519", keypair.PK_EDDSA},

	"ecdsa":   {"ecdsa", keypair.PK_ECDSA},
	"sm2":     {"sm2", keypair.PK_SM2},
	"ed25519": {"ed25519", keypair.PK_EDDSA},
}

type CurveInfo struct {
	Name string
	Code byte
}

var CurveMap = map[string]CurveInfo{
	"":  {"P-256", keypair.P256},
	"1": {"P-224", keypair.P224},
	"2": {"P-256", keypair.P256},
	"3": {"P-384", keypair.P384},
	"4": {"P-521", keypair.P521},

	"P-224": {"P-224", keypair.P224},
	"P-256": {"P-256", keypair.P256},
	"P-384": {"P-384", keypair.P384},
	"P-521": {"P-521", keypair.P521},

	"SM2P256V1": {"SM2P256V1", keypair.SM2P256V1},
	"ED25519":   {"ED25519", keypair.ED25519},
}

type SchemeInfo struct {
	Name string
	Code s.SignatureScheme
}

var SchemeMap = map[string]SchemeInfo{
	"":  {"SHA256withECDSA", s.SHA256withECDSA},
	"1": {"SHA224withECDSA", s.SHA224withECDSA},
	"2": {"SHA256withECDSA", s.SHA256withECDSA},
	"3": {"SHA384withECDSA", s.SHA384withECDSA},
	"4": {"SHA512withEDDSA", s.SHA512withEDDSA},
	"5": {"SHA3_224withECDSA", s.SHA3_224withECDSA},
	"6": {"SHA3_256withECDSA", s.SHA3_256withECDSA},
	"7": {"SHA3_384withECDSA", s.SHA3_384withECDSA},
	"8": {"SHA3_512withECDSA", s.SHA3_512withECDSA},
	"9": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SHA224withECDSA":    {"SHA224withECDSA", s.SHA224withECDSA},
	"SHA256withECDSA":    {"SHA256withECDSA", s.SHA256withECDSA},
	"SHA384withECDSA":    {"SHA384withECDSA", s.SHA384withECDSA},
	"SHA512withEDDSA":    {"SHA512withEDDSA", s.SHA512withEDDSA},
	"SHA3_224withECDSA":  {"SHA3_224withECDSA", s.SHA3_224withECDSA},
	"SHA3_256withECDSA":  {"SHA3_256withECDSA", s.SHA3_256withECDSA},
	"SHA3_384withECDSA":  {"SHA3_384withECDSA", s.SHA3_384withECDSA},
	"SHA3_512withECDSA":  {"SHA3_512withECDSA", s.SHA3_512withECDSA},
	"RIPEMD160withECDSA": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SM3withSM2": {"SM3withSM2", s.SM3withSM2},
}
