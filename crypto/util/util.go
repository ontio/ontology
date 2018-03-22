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

package util

import (
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"github.com/Ontology/crypto/sm3"
	//"math/big"
)

const (
	HASHLEN = 32
	PRIVATEKEYLEN = 32
	PUBLICKEYLEN = 32
	SIGNRLEN = 32
	SIGNSLEN = 32
	SIGNATURELEN = 64
	NEGBIGNUMLEN = 33
)

type CryptoAlgSet struct {
	EccParams elliptic.CurveParams
	Curve     elliptic.Curve
}

// RandomNum Generate the "real" random number which can be used for crypto algorithm
func RandomNum(n int) ([]byte, error) {
	// TODO Get the random number from System urandom
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}
	return b, nil
}

func Hash(data []byte) [HASHLEN]byte {
	return sha256.Sum256(data)
}

func SM3(data []byte) [HASHLEN]byte {
	return sm3.Sum(data)
}

// CheckMAC reports whether messageMAC is a valid HMAC tag for message.
func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func RIPEMD160(value []byte) []byte {
	//TODO: implement RIPEMD160

	return nil
}
