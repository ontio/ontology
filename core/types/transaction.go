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
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	sysconfig "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
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

func TransactionFromEIP155(eiptx *types.Transaction) (*Transaction, error) {
	evmChainId := sysconfig.DefConfig.P2PNode.EVMChainId

	signer := types.NewEIP155Signer(big.NewInt(int64(evmChainId)))
	from, err := signer.Sender(eiptx)
	if err != nil {
		return nil, fmt.Errorf("error EIP155 get sender:%s", err.Error())
	}

	addr := common.Address(from)

	if eiptx.Nonce() > uint64(math.MaxUint32) || !eiptx.GasPrice().IsUint64() {
		return nil, fmt.Errorf("nonce :%d or GasPrice :%d is too big", eiptx.Nonce(), eiptx.GasPrice())
	}

	retTx := &Transaction{
		Version:              byte(0),
		TxType:               EIP155,
		Nonce:                uint32(eiptx.Nonce()),
		GasPrice:             eiptx.GasPrice().Uint64(),
		GasLimit:             eiptx.Gas(),
		Payer:                addr,
		Payload:              &payload.EIP155Code{EIPTx: eiptx},
		hashUnsigned:         common.Uint256(signer.Hash(eiptx)),
		hash:                 common.Uint256(eiptx.Hash()),
		SignedAddr:           []common.Address{addr},
		nonDirectConstracted: true,
	}

	//raw = version + txtype + rlp(ethtx)
	raw, err := rlp.EncodeToBytes(eiptx)
	if err != nil {
		return nil, fmt.Errorf("error EIP155 EncodeToBytes %s", err.Error())
	}
	sink := new(common.ZeroCopySink)
	sink.WriteByte(retTx.Version)
	sink.WriteByte(byte(retTx.TxType))
	sink.WriteVarBytes(raw)

	retTx.Raw = sink.Bytes()

	return retTx, nil
}

func (tx *Transaction) GetEIP155Tx() (*types.Transaction, error) {
	if tx.TxType == EIP155 {
		tx := tx.Payload.(*payload.EIP155Code).EIPTx
		return tx, nil
	}
	return nil, fmt.Errorf("not a EIP155 tx")
}

func isEip155TxBytes(source *common.ZeroCopySource) bool {
	prefix, eof := source.NextBytes(2)
	if eof {
		return false
	}
	source.BackUp(2)
	return TransactionType(prefix[1]) == EIP155
}

func (tx *Transaction) decodeEip155(source *common.ZeroCopySource) error {
	pstart := source.Pos()
	tx.Version, _ = source.NextByte()
	txtype, eof := source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	tx.TxType = TransactionType(txtype)
	if tx.TxType != EIP155 {
		return fmt.Errorf("unreachable code path")
	}

	pl := new(payload.EIP155Code)
	err := pl.Deserialization(source)
	if err != nil {
		return err
	}
	decoded, err := TransactionFromEIP155(pl.EIPTx)
	if err != nil {
		return err
	}
	*tx = *decoded

	pend := source.Pos()
	lenAll := pend - pstart
	if lenAll > MAX_TX_SIZE {
		return fmt.Errorf("execced max transaction size:%d", lenAll)
	}

	return nil
}

// Transaction has internal reference of param `source`
func (tx *Transaction) Deserialization(source *common.ZeroCopySource) error {
	if isEip155TxBytes(source) {
		return tx.decodeEip155(source)
	}

	pstart := source.Pos()
	err := tx.deserializeOntUnsigned(source)
	if err != nil {
		return err
	}
	pos := source.Pos()
	lenUnsigned := pos - pstart
	source.BackUp(lenUnsigned)
	rawUnsigned, _ := source.NextBytes(lenUnsigned)

	tx.hashUnsigned = sha256.Sum256(rawUnsigned)
	tx.hash = sha256.Sum256(tx.hashUnsigned[:])

	// tx sigs
	length, err := source.ReadVarUint()
	if err != nil {
		return err
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

func (tx *Transaction) Value() *big.Int {
	if tx.TxType != EIP155 {
		return big.NewInt(0)
	}
	eiptx, err := tx.GetEIP155Tx()
	if err != nil {
		log.Error("GetEIP155Tx failed:%s", err.Error())
		return big.NewInt(0)
	}
	return eiptx.Value()
}

func (tx *Transaction) Cost() *big.Int {
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

func (tx *Transaction) deserializeOntUnsigned(source *common.ZeroCopySource) error {
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

	if tx.TxType == EIP155 {
		return fmt.Errorf("unreachable code path")
	}
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
	if tx.TxType == EIP155 {
		eiptx, err := tx.GetEIP155Tx()
		if err != nil {
			panic(err)
		}
		signer := types.NewEIP155Signer(big.NewInt(int64(id)))
		return common.Uint256(signer.Hash(eiptx))
	}

	sink := common.NewZeroCopySink(nil)
	sink.WriteHash(tx.hashUnsigned)
	if id != 0 {
		sink.WriteUint32(id)
	}

	return sha256.Sum256(sink.Bytes())
}
