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
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/program"
	"github.com/ontio/ontology/errors"
)

type Transaction struct {
	Version  byte
	TxType   TransactionType
	Nonce    uint32
	GasPrice uint64
	GasLimit uint64
	Payer    common.Address
	Payload  Payload
	//Attributes []*TxAttribute
	attributes byte //this must be 0 now, Attribute Array length use VarUint encoding, so byte is enough for extention
	Sigs       []*Sig

	hash *common.Uint256
}

type Sig struct {
	SigData [][]byte
	PubKeys []keypair.PublicKey
	M       uint16
}

func (self *Sig) Serialize(w io.Writer) error {
	invocationScript := program.ProgramFromParams(self.SigData)
	var verificationScript []byte
	if len(self.PubKeys) == 0 {
		return errors.NewErr("no pubkeys in sig")
	} else if len(self.PubKeys) == 1 {
		verificationScript = program.ProgramFromPubKey(self.PubKeys[0])
	} else {
		script, err := program.ProgramFromMultiPubKey(self.PubKeys, int(self.M))
		if err != nil {
			return err
		}
		verificationScript = script
	}
	err := serialization.WriteVarBytes(w, invocationScript)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, verificationScript)
	if err != nil {
		return err
	}

	return nil
}

func (self *Sig) Deserialize(r io.Reader) error {
	invocationScript, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	verificationScript, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	sigs, err := program.GetParamInfo(invocationScript)
	if err != nil {
		return err
	}
	info, err := program.GetProgramInfo(verificationScript)
	if err != nil {
		return err
	}

	self.SigData = sigs
	self.M = info.M
	self.PubKeys = info.PubKeys

	return nil
}

func (self *Transaction) GetSignatureAddresses() []common.Address {
	address := make([]common.Address, 0, len(self.Sigs))
	for _, sig := range self.Sigs {
		m := int(sig.M)
		n := len(sig.PubKeys)

		if n == 1 {
			address = append(address, AddressFromPubKey(sig.PubKeys[0]))
		} else {
			addr, err := AddressFromMultiPubKeys(sig.PubKeys, m)
			if err != nil {
				return nil
			}
			address = append(address, addr)
		}
	}
	return address
}

type TransactionType byte

const (
	Bookkeeper TransactionType = 0x02
	Deploy     TransactionType = 0xd0
	Invoke     TransactionType = 0xd1
)

// Payload define the func for loading the payload data
// base on payload type which have different struture
type Payload interface {

	//Serialize payload data
	Serialize(w io.Writer) error

	Deserialize(r io.Reader) error
}

// Serialize the Transaction
func (tx *Transaction) Serialize(w io.Writer) error {

	err := tx.SerializeUnsigned(w)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Serialize], Transaction txSerializeUnsigned Serialize failed.")
	}

	err = serialization.WriteVarUint(w, uint64(len(tx.Sigs)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Serialize], Transaction serialize tx sigs length failed.")
	}
	for _, sig := range tx.Sigs {
		err = sig.Serialize(w)
		if err != nil {
			return err
		}
	}

	return nil
}

//Serialize the Transaction data without contracts
func (tx *Transaction) SerializeUnsigned(w io.Writer) error {
	//txType
	if _, err := w.Write([]byte{byte(tx.Version), byte(tx.TxType)}); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction version failed.")
	}
	if err := serialization.WriteUint32(w, tx.Nonce); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction nonce failed.")
	}
	if err := serialization.WriteUint64(w, tx.GasPrice); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction gasPrice failed.")
	}
	if err := serialization.WriteUint64(w, tx.GasLimit); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction gasLimit failed.")
	}
	if err := tx.Payer.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction payer failed.")
	}
	//Payload
	if tx.Payload == nil {
		return errors.NewErr("Transaction Payload is nil.")
	}
	if err := tx.Payload.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction payload failed.")
	}

	err := serialization.WriteVarUint(w, uint64(tx.attributes))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SerializeUnsigned], Transaction item txAttribute length serialization failed.")
	}

	return nil
}

// deserialize the Transaction
func (tx *Transaction) Deserialize(r io.Reader) error {
	// tx deserialize
	err := tx.DeserializeUnsigned(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Deserialize], Transaction deserializeUnsigned error.")
	}

	// tx sigs
	length, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Deserialize], Transaction sigs length deserialize error.")
	}

	for i := 0; i < int(length); i++ {
		sig := new(Sig)
		err := sig.Deserialize(r)
		if err != nil {
			return errors.NewErr("deserialize transaction failed")
		}
		tx.Sigs = append(tx.Sigs, sig)
	}

	return nil
}

func (tx *Transaction) DeserializeUnsigned(r io.Reader) error {
	var versiontype [2]byte
	_, err := io.ReadFull(r, versiontype[:])
	if err != nil {
		return err
	}
	nonce, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	gasPrice, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	gasLimit, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	tx.Version = versiontype[0]
	tx.TxType = TransactionType(versiontype[1])
	tx.Nonce = nonce
	tx.GasPrice = gasPrice
	tx.GasLimit = gasLimit
	if err := tx.Payer.Deserialize(r); err != nil {
		return err
	}

	switch tx.TxType {
	case Invoke:
		tx.Payload = new(payload.InvokeCode)
	case Deploy:
		tx.Payload = new(payload.DeployCode)
	default:
		return errors.NewErr(fmt.Sprintf("unsupported tx type %v", tx.Type()))
	}

	err = tx.Payload.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[DeserializeUnsigned], Transaction payload parse error.")
	}

	//attributes
	length, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}
	if length != 0 {
		return fmt.Errorf("transaction attribute must be 0, got %d", length)
	}
	tx.attributes = 0

	return nil
}

func (tx *Transaction) GetMessage() []byte {
	buf := new(bytes.Buffer)
	tx.SerializeUnsigned(buf)
	return buf.Bytes()
}

func (tx *Transaction) ToArray() []byte {
	b := new(bytes.Buffer)
	tx.Serialize(b)
	return b.Bytes()
}

func (tx *Transaction) Hash() common.Uint256 {
	if tx.hash == nil {
		buf := bytes.Buffer{}
		tx.SerializeUnsigned(&buf)

		temp := sha256.Sum256(buf.Bytes())
		f := common.Uint256(sha256.Sum256(temp[:]))
		tx.hash = &f
	}
	return *tx.hash
}

func (tx *Transaction) SetHash(hash common.Uint256) {
	tx.hash = &hash
}

func (tx *Transaction) Type() common.InventoryType {
	return common.TRANSACTION
}

func (tx *Transaction) Verify() error {
	panic("unimplemented ")
	return nil
}
