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

package signature

import (
	"errors"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
)

// Sign returns the signature of data using privKey
func Sign(signer Signer, data []byte) ([]byte, error) {
	signature, err := s.Sign(signer.Scheme(), signer.PrivKey(), data, nil)
	if err != nil {
		return nil, err
	}

	return s.Serialize(signature)
}

// Verify check the signature of data using pubKey
func Verify(pubKey keypair.PublicKey, data, signature []byte) error {
	sigObj, err := s.Deserialize(signature)
	if err != nil {
		return errors.New("invalid signature data: " + err.Error())
	}

	if !s.Verify(pubKey, data, sigObj) {
		return errors.New("signature verification failed")
	}

	return nil
}

// VerifyMultiSignature check whether more than m sigs are signed by the keys
func VerifyMultiSignature(data []byte, keys []keypair.PublicKey, m int, sigs [][]byte) error {
	n := len(keys)

	if len(sigs) < m {
		return errors.New("not enough signatures in multi-signature")
	}

	mask := make([]bool, n)
	for i := 0; i < m; i++ {
		valid := false

		sig, err := s.Deserialize(sigs[i])
		if err != nil {
			return errors.New("invalid signature data")
		}
		for j := 0; j < n; j++ {
			if mask[j] {
				continue
			}
			if s.Verify(keys[j], data, sig) {
				mask[j] = true
				valid = true
				break
			}
		}

		if valid == false {
			return errors.New("multi-signature verification failed")
		}
	}

	return nil
}
