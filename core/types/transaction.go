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
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/ontology/common/log"
	"io"
	"math/big"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	sysconfig "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/program"
)

const MAX_TX_SIZE = 1024 * 1024 // The max size of a transaction to prevent DOS attacks

type Transaction struct {
	Version  byte
	TxType   TransactionType
	Nonce    uint32
	GasPrice uint64
	GasLimit uint64
	Payer    common.Address
	Payload  Payload
	//Attributes []*TxAttribute
	attributes byte //this must be 0 now, Attribute Array length use VarUint encoding, so byte is enough for extension
	Sigs       []RawSig

	Raw []byte // raw transaction data

	hashUnsigned common.Uint256
	hash         common.Uint256
	SignedAddr   []common.Address // this is assigned when passed signature verification

	nonDirectConstracted bool // used to check literal construction like `tx := &Transaction{...}`
}

// if no error, ownership of param raw is transfered to Transaction
func TransactionFromRawBytes(raw []byte) (*Transaction, error) {
	if len(raw) > MAX_TX_SIZE {
		return nil, errors.New("execced max transaction size")
	}
	source := common.NewZeroCopySource(raw)
	tx := &Transaction{Raw: raw}
	err := tx.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// todo transaction from EIP155 tx
func TransactionFromEIP155(eiptx *types.Transaction) (*Transaction, error) {

	evmChainId := sysconfig.DefConfig.P2PNode.EVMChainId

	signer := types.NewEIP155Signer(big.NewInt(int64(evmChainId)))
	from, err := signer.Sender(eiptx)
	if err != nil {
		return nil, fmt.Errorf("error EIP155 get sender:%s", err.Error())
	}

	addr, err := common.AddressParseFromBytes(from[:])
	if err != nil {
		return nil, fmt.Errorf("error EIP155 parse sender address:%s", err.Error())
	}
	eiphash := eiptx.Hash()
	txhash, err := common.Uint256ParseFromBytes(eiphash[:])
	if err != nil {
		return nil, fmt.Errorf("error EIP155 parse txhash:%s", err.Error())
	}
	retTx := &Transaction{
		Version:  byte(0),
		TxType:   EIP155,
		Nonce:    uint32(eiptx.Nonce()),
		GasPrice: eiptx.GasPrice().Uint64(),
		GasLimit: eiptx.Gas(),
		Payer:    addr,
		Payload:  &payload.EIP155Code{Code: eiptx.Data()},
		//Sigs: ???
		//Raw:eiptx.Data(),
		hashUnsigned:         common.Uint256{},
		hash:                 txhash,
		SignedAddr:           []common.Address{addr},
		nonDirectConstracted: true,
	}

	sink := new(common.ZeroCopySink)
	retTx.Serialization(sink)
	retTx.Raw = sink.Bytes()

	return retTx, nil
}

func(tx *Transaction)GetEIP155Tx()(*types.Transaction,error){
	if tx.TxType == EIP155 {
		bts := tx.Payload.(*payload.EIP155Code).Code
		eiptx := new(types.Transaction)
		err := eiptx.DecodeRLP(rlp.NewStream(bytes.NewBuffer(bts), uint64(len(bts))))
		if err != nil {
			return nil,fmt.Errorf("error on DecodeRLP :%s", err.Error())
		}
		return eiptx,nil
	}
	return nil,fmt.Errorf("not a EIP155 tx")
}


func (tx *Transaction) VerifyEIP155Tx() error {
	if tx.TxType != EIP155 {
		return fmt.Errorf("not a EIP155 transaction")
	}

	bts := tx.Payload.(*payload.EIP155Code).Code
	eiptx := new(types.Transaction)
	err := eiptx.DecodeRLP(rlp.NewStream(bytes.NewBuffer(bts), uint64(len(bts))))
	if err != nil {
		return fmt.Errorf("error on DecodeRLP :%s", err.Error())
	}
	v, r, s := eiptx.RawSignatureValues()
	return sanityCheckSignature(v, r, s, true)
}

func sanityCheckSignature(v *big.Int, r *big.Int, s *big.Int, maybeProtected bool) error {
	if isProtectedV(v) && !maybeProtected {
		return errors.New("transaction type does not supported EIP-155 protected signatures")
	}

	var plainV byte
	if isProtectedV(v) {
		chainID := deriveChainId(v).Uint64()
		plainV = byte(v.Uint64() - 35 - 2*chainID)
	} else if maybeProtected {
		// Only EIP-155 signatures can be optionally protected. Since
		// we determined this v value is not protected, it must be a
		// raw 27 or 28.
		plainV = byte(v.Uint64() - 27)
	} else {
		// If the signature is not optionally protected, we assume it
		// must already be equal to the recovery id.
		plainV = byte(v.Uint64())
	}
	if !crypto.ValidateSignatureValues(plainV, r, s, false) {
		return errors.New("transaction type not valid in this context")
	}

	return nil
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28 && v != 1 && v != 0
	}
	// anything not 27 or 28 is considered protected
	return true
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

// Transaction has internal reference of param `source`
func (tx *Transaction) Deserialization(source *common.ZeroCopySource) error {

	pstart := source.Pos()
	err := tx.deserializationUnsigned(source)
	if err != nil {
		return err
	}
	pos := source.Pos()
	lenUnsigned := pos - pstart
	source.BackUp(lenUnsigned)
	rawUnsigned, _ := source.NextBytes(lenUnsigned)

	tx.hashUnsigned = sha256.Sum256(rawUnsigned)
	//todo deal with EIP155 tx hash
	if tx.TxType == EIP155 {
		//payload is EIP155 bytes
		sink := common.NewZeroCopySink(nil)
		tx.Payload.Serialization(sink)
		bts := sink.Bytes()

		eiptx := new(types.Transaction)
		err := eiptx.DecodeRLP(rlp.NewStream(bytes.NewBuffer(bts), uint64(len(bts))))
		if err != nil {
			return fmt.Errorf("error on DecodeRLP :%s", err.Error())
		}
		eiphash := eiptx.Hash()
		tx.hash, err = common.Uint256ParseFromBytes(eiphash[:])
		if err != nil {
			return fmt.Errorf("error EIP155 parse txhash:%s", err.Error())
		}
	} else {
		tx.hash = sha256.Sum256(tx.hashUnsigned[:])
	}

	// tx sigs
	length, _, irregular, eof := source.NextVarUint()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	if length > constants.TX_MAX_SIG_SIZE {
		return fmt.Errorf("transaction signature number %d execced %d", length, constants.TX_MAX_SIG_SIZE)
	}

	for i := 0; i < int(length); i++ {
		var sig RawSig
		err := sig.Deserialization(source)
		if err != nil {
			return err
		}

		tx.Sigs = append(tx.Sigs, sig)
	}

	pend := source.Pos()
	lenAll := pend - pstart
	if lenAll > MAX_TX_SIZE {
		return fmt.Errorf("execced max transaction size:%d", lenAll)
	}
	source.BackUp(lenAll)
	tx.Raw, _ = source.NextBytes(lenAll)

	tx.nonDirectConstracted = true

	return nil
}

func (tx *Transaction)Value() *big.Int  {
	eiptx,err := tx.GetEIP155Tx()
	if err != nil {
		log.Error("GetEIP155Tx failed:%s",err.Error())
		return big.NewInt(0)
	}
	return eiptx.Value()
}

func  (tx *Transaction)Cost() *big.Int{
	total := new(big.Int).Mul(new(big.Int).SetUint64(tx.GasPrice), new(big.Int).SetUint64(tx.GasLimit))
	total.Add(total, tx.Value())
	return total
}

// note: ownership transfered to output
func (tx *Transaction) IntoMutable() (*MutableTransaction, error) {
	mutable := &MutableTransaction{
		Version:  tx.Version,
		TxType:   tx.TxType,
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Payer:    tx.Payer,
		Payload:  tx.Payload,
	}

	for _, raw := range tx.Sigs {
		sig, err := raw.GetSig()
		if err != nil {
			return nil, err
		}
		mutable.Sigs = append(mutable.Sigs, sig)
	}

	return mutable, nil
}

func (tx *Transaction) deserializationUnsigned(source *common.ZeroCopySource) error {
	var irregular, eof bool
	tx.Version, eof = source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	var txtype byte
	txtype, eof = source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	tx.TxType = TransactionType(txtype)
	tx.Nonce, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	tx.GasPrice, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	tx.GasLimit, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	var buf []byte
	buf, eof = source.NextBytes(common.ADDR_LEN)
	if eof {
		return io.ErrUnexpectedEOF
	}
	copy(tx.Payer[:], buf)

	switch tx.TxType {
	case InvokeNeo, InvokeWasm:
		pl := new(payload.InvokeCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl
	case Deploy:
		pl := new(payload.DeployCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl
	case EIP155:
		pl := new(payload.EIP155Code)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl

	default:
		return fmt.Errorf("unsupported tx type %v", tx.TxType)
	}

	var length uint64
	length, _, irregular, eof = source.NextVarUint()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	if length != 0 {
		return fmt.Errorf("transaction attribute must be 0, got %d", length)
	}
	tx.attributes = 0

	return nil
}

type RawSig struct {
	Invoke []byte
	Verify []byte
}

func (self *RawSig) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(self.Invoke)
	sink.WriteVarBytes(self.Verify)
	return nil
}

func (self *RawSig) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	self.Invoke, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	self.Verify, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

type Sig struct {
	SigData [][]byte
	PubKeys []keypair.PublicKey
	M       uint16
}

func (self *Sig) GetRawSig() (*RawSig, error) {
	invocationScript := program.ProgramFromParams(self.SigData)
	var verificationScript []byte
	if len(self.PubKeys) == 0 {
		return nil, errors.New("no pubkeys in sig")
	} else if len(self.PubKeys) == 1 {
		verificationScript = program.ProgramFromPubKey(self.PubKeys[0])
	} else {
		script, err := program.ProgramFromMultiPubKey(self.PubKeys, int(self.M))
		if err != nil {
			return nil, err
		}
		verificationScript = script
	}

	return &RawSig{Invoke: invocationScript, Verify: verificationScript}, nil
}

func (self *RawSig) GetSig() (Sig, error) {
	sigs, err := program.GetParamInfo(self.Invoke)
	if err != nil {
		return Sig{}, err
	}
	info, err := program.GetProgramInfo(self.Verify)
	if err != nil {
		return Sig{}, err
	}

	return Sig{SigData: sigs, M: info.M, PubKeys: info.PubKeys}, nil
}

func (self *Sig) Serialization(sink *common.ZeroCopySink) error {
	temp := common.NewZeroCopySink(nil)
	program.EncodeParamProgramInto(temp, self.SigData)
	sink.WriteVarBytes(temp.Bytes())

	temp.Reset()
	if len(self.PubKeys) == 0 {
		return errors.New("no pubkeys in sig")
	} else if len(self.PubKeys) == 1 {
		program.EncodeSinglePubKeyProgramInto(temp, self.PubKeys[0])
	} else {
		err := program.EncodeMultiPubKeyProgramInto(temp, self.PubKeys, int(self.M))
		if err != nil {
			return err
		}
	}
	sink.WriteVarBytes(temp.Bytes())

	return nil
}

func (self *Transaction) GetSignatureAddresses() []common.Address {
	if len(self.SignedAddr) == 0 {
		addrs := make([]common.Address, 0, len(self.Sigs))
		for _, prog := range self.Sigs {
			addrs = append(addrs, common.AddressFromVmCode(prog.Verify))
		}
		self.SignedAddr = addrs
	}
	return self.SignedAddr
}

type TransactionType byte

const (
	Bookkeeper TransactionType = 0x02
	Deploy     TransactionType = 0xd0
	InvokeNeo  TransactionType = 0xd1
	InvokeWasm TransactionType = 0xd2 //add for wasm invoke
	EIP155     TransactionType = 0xd3 //for EIP155 transaction
)

// Payload define the func for loading the payload data
// base on payload type which have different structure
type Payload interface {
	//Serialize payload data
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

func (tx *Transaction) Serialization(sink *common.ZeroCopySink) {
	if !tx.nonDirectConstracted || len(tx.Raw) == 0 {
		panic("wrong constructed transaction")
	}
	sink.WriteBytes(tx.Raw)
}

func (tx *Transaction) ToArray() []byte {
	return common.SerializeToBytes(tx)
}

func (tx *Transaction) Hash() common.Uint256 {
	return tx.hash
}

// calculate a hash for another chain to sign.
// and take the chain id of ontology as 0.
func (tx *Transaction) SigHashForChain(id uint32) common.Uint256 {
	sink := common.NewZeroCopySink(nil)
	sink.WriteHash(tx.hashUnsigned)
	if id != 0 {
		sink.WriteUint32(id)
	}

	return sha256.Sum256(sink.Bytes())
}
