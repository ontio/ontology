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
	"bytes"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/core/contract/program"
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
)

//SignableData describe the data need be signed.
type SignableData interface {
	interfaces.CodeContainer

	////Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Address, error)

	SetPrograms([]*program.Program)

	GetPrograms() []*program.Program

	//TODO: add SerializeUnsigned
	SerializeUnsigned(io.Writer) error
}

// Default signature scheme
var defaultScheme = s.SHA256withECDSA

func SetDefaultScheme(scheme string) error {
	res, err := s.GetScheme(scheme)
	if err != nil {
		return errors.New(err.Error() + ", use SHA256withECDSA as default")
	}
	defaultScheme = res

	return nil
}

func SignBySigner(data SignableData, signer Signer) ([]byte, error) {
	return Sign(signer.PrivKey(), getHashData(data))
}

func getHashData(data SignableData) []byte {
	buf := new(bytes.Buffer)
	data.SerializeUnsigned(buf)
	temp := sha256.Sum256(buf.Bytes())
	hash := sha256.Sum256(temp[:])
	return hash[:]
}

func Sign(privKey keypair.PrivateKey, data []byte) ([]byte, error) {
	signature, err := s.Sign(defaultScheme, privKey, data, nil)
	if err != nil {
		return nil, err
	}

	return s.Serialize(signature)
}

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
