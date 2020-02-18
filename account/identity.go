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
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/itchyny/base58-go"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/types"
	"golang.org/x/crypto/ripemd160"
)

const (
	SCHEME = "did"
	METHOD = "ont"
	VER    = 0x41
)

func GenerateID() (string, error) {
	var buf [32]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return "", fmt.Errorf("generate ID error, %s", err)
	}
	return CreateID(buf[:])
}

func CreateID(nonce []byte) (string, error) {
	hasher := ripemd160.New()
	_, err := hasher.Write(nonce)
	if err != nil {
		return "", fmt.Errorf("create ID error, %s", err)
	}
	data := hasher.Sum([]byte{VER})
	data = append(data, checksum(data)...)

	bi := new(big.Int).SetBytes(data).String()
	idstring, err := base58.BitcoinEncoding.Encode([]byte(bi))
	if err != nil {
		return "", fmt.Errorf("create ID error, %s", err)
	}

	return SCHEME + ":" + METHOD + ":" + string(idstring), nil
}

func VerifyID(id string) bool {
	if len(id) < 9 {
		return false
	}
	if id[0:8] != "did:ont:" {
		return false
	}
	buf, err := base58.BitcoinEncoding.Decode([]byte(id[8:]))
	if err != nil {
		return false
	}
	bi, ok := new(big.Int).SetString(string(buf), 10)
	if !ok || bi == nil {
		return false
	}
	buf = bi.Bytes()
	// 1 byte version + 20 byte hash + 4 byte checksum
	if len(buf) != 25 {
		return false
	}
	pos := len(buf) - 4
	data := buf[:pos]
	check := buf[pos:]
	sum := checksum(data)

	return bytes.Equal(sum, check)
}

func checksum(data []byte) []byte {
	sum := sha256.Sum256(data)
	sum = sha256.Sum256(sum[:])
	return sum[:4]
}

type Identity struct {
	ID      string       `json:"ontid"`
	Label   string       `json:"label,omitempty"`
	Lock    bool         `json:"lock"`
	Control []Controller `json:"controls,omitempty"`
	Extra   interface{}  `json:"extra,omitempty"`
}

type Controller struct {
	ID     string `json:"id"`
	Public string `json:"publicKey,omitempty"`
	keypair.ProtectedKey
}

func NewIdentity(label string, keyType keypair.KeyType, param interface{}, password []byte) (*Identity, error) {
	id, err := GenerateID()
	if err != nil {
		return nil, err
	}

	pri, pub, err := keypair.GenerateKeyPair(keyType, param)
	if err != nil {
		return nil, err
	}
	addr := types.AddressFromPubKey(pub)
	b58addr := addr.ToBase58()
	prot, err := keypair.EncryptPrivateKey(pri, b58addr, password)
	if err != nil {
		return nil, err
	}

	var res Identity
	res.ID = id
	res.Label = label
	res.Lock = false
	res.Control = make([]Controller, 1)
	res.Control[0].ID = "1"
	res.Control[0].ProtectedKey = *prot
	res.Control[0].Public = hex.EncodeToString(keypair.SerializePublicKey(pub))

	return &res, nil
}
