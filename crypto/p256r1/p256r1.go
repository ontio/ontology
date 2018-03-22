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

package p256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Ontology/crypto/util"
	"math/big"
)

func Init(algSet *util.CryptoAlgSet) {
	algSet.Curve = elliptic.P256()
	algSet.EccParams = *(algSet.Curve.Params())
}

func GenKeyPair(algSet *util.CryptoAlgSet) ([]byte, *big.Int, *big.Int, error) {
	privateKey := new(ecdsa.PrivateKey)
	privateKey, err := ecdsa.GenerateKey(algSet.Curve, rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.New("Generate key pair error")
	}

	priKey := privateKey.D.Bytes()
	return priKey, privateKey.PublicKey.X, privateKey.PublicKey.Y, nil
}

func Sign(algSet *util.CryptoAlgSet, priKey []byte, data []byte) (*big.Int, *big.Int, error) {
	digest := util.Hash(data)

	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = algSet.Curve
	privateKey.D = big.NewInt(0)
	privateKey.D.SetBytes(priKey)

	r := big.NewInt(0)
	s := big.NewInt(0)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest[:])
	if err != nil {
		fmt.Printf("Sign error\n")
		return nil, nil, err
	}
	return r, s, nil
}

func Verify(algSet *util.CryptoAlgSet, X *big.Int, Y *big.Int, data []byte, r, s *big.Int) error {
	digest := util.Hash(data)

	pub := new(ecdsa.PublicKey)
	pub.Curve = algSet.Curve

	pub.X = new(big.Int).Set(X)
	pub.Y = new(big.Int).Set(Y)

	if ecdsa.Verify(pub, digest[:], r, s) {
		return nil
	} else {
		return errors.New("[Validation], Verify failed.")
	}
}
