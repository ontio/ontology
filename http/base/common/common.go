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

package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/smartcontract/event"
	cstates "github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"math/big"
	"strings"
	"time"
)

type BalanceOfRsp struct {
	Ont string `json:"ont"`
	Ong string `json:"ong"`
}

type MerkleProof struct {
	Type             string
	TransactionsRoot string
	BlockHeight      uint32
	CurBlockRoot     string
	CurBlockHeight   uint32
	TargetHashes     []string
}

type ExecuteNotify struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []NotifyEventInfo
}

type NotifyEventInfo struct {
	ContractAddress string
	States          interface{}
}

type TxAttributeInfo struct {
	Usage types.TransactionAttributeUsage
	Data  string
}

type AmountMap struct {
	Key   common.Uint256
	Value common.Fixed64
}

type Fee struct {
	Amount common.Fixed64
	Payer  string
}

type Sig struct {
	PubKeys []string
	M       uint8
	SigData []string
}
type Transactions struct {
	Version    byte
	Nonce      uint32
	GasPrice   uint64
	GasLimit   uint64
	Payer      string
	TxType     types.TransactionType
	Payload    PayloadInfo
	Attributes []TxAttributeInfo
	Sigs       []Sig
	Hash       string
}

type BlockHead struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	BlockRoot        string
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload string
	NextBookkeeper   string

	Bookkeepers []string
	SigData     []string

	Hash string
}

type BlockInfo struct {
	Hash         string
	Header       *BlockHead
	Transactions []*Transactions
}

type NodeInfo struct {
	NodeState   uint   // node status
	NodePort    uint16 // The nodes's port
	ID          uint64 // The nodes's id
	NodeTime    int64
	NodeVersion uint32   // The network protocol the node used
	NodeType    uint64   // The services the node supplied
	Relay       bool     // The relay capability of the node (merge into capbility flag)
	Height      uint32   // The node latest block height
	TxnCnt      []uint64 // The transactions be transmit by this node
	//RxTxnCnt uint64 // The transaction received by this node
}

type ConsensusInfo struct {
	// TODO
}

type TXNAttrInfo struct {
	Height  uint32
	Type    int
	ErrCode int
}

type TXNEntryInfo struct {
	Txn   Transactions  // transaction which has been verified
	Fee   int64         // Total fee per transaction
	Attrs []TXNAttrInfo // the result from each validator
}

func GetExecuteNotify(obj *event.ExecuteNotify) (map[string]bool, ExecuteNotify) {
	evts := []NotifyEventInfo{}
	var contractAddrs = make(map[string]bool)
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.ContractAddress.ToHexString(), v.States})
		contractAddrs[v.ContractAddress.ToHexString()] = true
	}
	txhash := common.ToHexString(obj.TxHash[:])
	return contractAddrs, ExecuteNotify{txhash, obj.State, obj.GasConsumed, evts}
}

func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.Nonce = ptx.Nonce
	trans.GasLimit = ptx.GasLimit
	trans.GasPrice = ptx.GasPrice
	trans.Payer = ptx.Payer.ToHexString()
	trans.Payload = TransPayloadToHex(ptx.Payload)

	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
	for i, v := range ptx.Attributes {
		trans.Attributes[i].Usage = v.Usage
		trans.Attributes[i].Data = common.ToHexString(v.Data)
	}
	trans.Sigs = []Sig{}
	for _, sig := range ptx.Sigs {
		e := Sig{M: sig.M}
		for i := 0; i < len(sig.PubKeys); i++ {
			key := keypair.SerializePublicKey(sig.PubKeys[i])
			e.PubKeys = append(e.PubKeys, common.ToHexString(key))
		}
		for i := 0; i < len(sig.SigData); i++ {
			e.SigData = append(e.SigData, common.ToHexString(sig.SigData[i]))
		}
		trans.Sigs = append(trans.Sigs, e)
	}

	mhash := ptx.Hash()
	trans.Hash = common.ToHexString(mhash.ToArray())
	return trans
}

func VerifyAndSendTx(txn *types.Transaction) ontErrors.ErrCode {
	// if transaction is verified unsuccessfully then will not put it into transaction pool
	if errCode := bactor.AppendTxToPool(txn); errCode != ontErrors.ErrNoError {
		log.Warn("Can NOT add the transaction to TxnPool")
		return errCode
	}
	return ontErrors.ErrNoError
}

func GetBlockInfo(block *types.Block) BlockInfo {
	hash := block.Hash()
	var bookkeepers = []string{}
	var sigData = []string{}
	for i := 0; i < len(block.Header.SigData); i++ {
		s := common.ToHexString(block.Header.SigData[i])
		sigData = append(sigData, s)
	}
	for i := 0; i < len(block.Header.Bookkeepers); i++ {
		e := block.Header.Bookkeepers[i]
		key := keypair.SerializePublicKey(e)
		bookkeepers = append(bookkeepers, common.ToHexString(key))
	}

	blockHead := &BlockHead{
		Version:          block.Header.Version,
		PrevBlockHash:    common.ToHexString(block.Header.PrevBlockHash.ToArray()),
		TransactionsRoot: common.ToHexString(block.Header.TransactionsRoot.ToArray()),
		BlockRoot:        common.ToHexString(block.Header.BlockRoot.ToArray()),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		ConsensusPayload: common.ToHexString(block.Header.ConsensusPayload),
		NextBookkeeper:   block.Header.NextBookkeeper.ToBase58(),
		Bookkeepers:      bookkeepers,
		SigData:          sigData,
		Hash:             common.ToHexString(hash.ToArray()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         common.ToHexString(hash.ToArray()),
		Header:       blockHead,
		Transactions: trans,
	}
	return b
}

func GetBalance(address common.Address) (*BalanceOfRsp, error) {
	ont, err := GetContractBalance(0, genesis.OntContractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get ont balance error:%s", err)
	}
	ong, err := GetContractBalance(0, genesis.OngContractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get ont balance error:%s", err)
	}
	return &BalanceOfRsp{
		Ont: fmt.Sprintf("%d", ont),
		Ong: fmt.Sprintf("%d", ong),
	}, nil
}

func GetAllowance(asset string, from, to common.Address) (string, error) {
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case "ont":
		contractAddr = genesis.OntContractAddress
	case "ong":
		contractAddr = genesis.OngContractAddress
	default:
		return "", fmt.Errorf("unsupport asset")
	}
	allowance, err := GetContractAllowance(0, contractAddr, from, to)
	if err != nil {
		return "", fmt.Errorf("get allowance error:%s", err)
	}
	return fmt.Sprintf("%v", allowance), nil
}

func GetContractBalance(cVersion byte, contractAddr, accAddr common.Address) (uint64, error) {
	addrBuf := bytes.NewBuffer(nil)
	err := accAddr.Serialize(addrBuf)
	if err != nil {
		return 0, fmt.Errorf("address serialize error:%s", err)
	}
	argBuf := bytes.NewBuffer(nil)
	err = serialization.WriteVarBytes(argBuf, addrBuf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("serialization.WriteVarBytes error:%s", err)
	}
	crt := &cstates.Contract{
		Version: cVersion,
		Address: contractAddr,
		Method:  "balanceOf",
		Args:    argBuf.Bytes(),
	}
	buf := bytes.NewBuffer(nil)
	err = crt.Serialize(buf)
	if err != nil {
		return 0, fmt.Errorf("Serialize contract error:%s", err)
	}
	result, err := PrepareInvokeContract(cVersion, vmtypes.Native, buf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("PrepareInvokeContract error:%s", err)
	}
	if result.State == 0 {
		return 0, fmt.Errorf("prepare invoke failed")
	}
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	balance := new(big.Int).SetBytes(data)
	return balance.Uint64(), nil
}

func GetContractAllowance(cVersion byte, contractAddr, fromAddr, toAddr common.Address) (uint64, error) {
	fromBuf := bytes.NewBuffer(nil)
	err := fromAddr.Serialize(fromBuf)
	if err != nil {
		return 0, fmt.Errorf("from address serialize error:%s", err)
	}
	toBuf := bytes.NewBuffer(nil)
	err = toAddr.Serialize(toBuf)
	if err != nil {
		return 0, fmt.Errorf("to address serialize error:%s", err)
	}

	argBuf := bytes.NewBuffer(nil)
	err = serialization.WriteVarBytes(argBuf, fromBuf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("serialization.WriteVarBytes error:%s", err)
	}
	err = serialization.WriteVarBytes(argBuf, toBuf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("serialization.WriteVarBytes error:%s", err)
	}
	crt := &cstates.Contract{
		Version: cVersion,
		Address: contractAddr,
		Method:  "allowance",
		Args:    argBuf.Bytes(),
	}
	buf := bytes.NewBuffer(nil)
	err = crt.Serialize(buf)
	if err != nil {
		return 0, fmt.Errorf("Serialize contract error:%s", err)
	}
	result, err := PrepareInvokeContract(cVersion, vmtypes.Native, buf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("PrepareInvokeContract error:%s", err)
	}
	if result.State == 0 {
		return 0, fmt.Errorf("prepare invoke failed")
	}
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	allowance := new(big.Int).SetBytes(data)
	return allowance.Uint64(), nil
}

func PrepareInvokeContract(cVersion byte, vmType vmtypes.VmType, invokeCode []byte) (*cstates.PreExecResult, error) {
	invokePayload := &payload.InvokeCode{
		Code: vmtypes.VmCode{
			VmType: vmType,
			Code:   invokeCode,
		},
	}
	tx := &types.Transaction{
		Version:    cVersion,
		TxType:     types.Invoke,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    invokePayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
		Sigs:       make([]*types.Sig, 0, 0),
	}
	return bactor.PreExecuteContract(tx)
}
