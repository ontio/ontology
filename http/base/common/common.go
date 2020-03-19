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

// Package common privides functions for http handler call
package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	common2 "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	cstate "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/vm/neovm"
)

const MAX_SEARCH_HEIGHT uint32 = 100
const MAX_REQUEST_BODY_SIZE = 1 << 20

type BalanceOfRsp struct {
	Ont    string `json:"ont"`
	Ong    string `json:"ong"`
	Height string `json:"height"`
}

type MerkleProof struct {
	Type             string
	TransactionsRoot string
	BlockHeight      uint32
	CurBlockRoot     string
	CurBlockHeight   uint32
	TargetHashes     []string
}

type LogEventArgs struct {
	TxHash          string
	ContractAddress string
	Message         string
}

type ExecuteNotify struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []NotifyEventInfo
}

type PreExecuteResult struct {
	State  byte
	Gas    uint64
	Result interface{}
	Notify []NotifyEventInfo
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
	M       uint16
	SigData []string
}

type CrossStatesProof struct {
	Type      string
	AuditPath string
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
	Height     uint32
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
	Size         int
	Header       *BlockHead
	Transactions []*Transactions
}

type NodeInfo struct {
	NodePort    uint16         // The nodes's port
	ID          common2.PeerId // The nodes's id
	NodeTime    int64
	NodeVersion uint32   // The network protocol the node used
	NodeType    uint64   // The services the node supplied
	Relay       bool     // The relay capability of the node (merge into capbility flag)
	Height      uint32   // The node latest block height
	TxnCnt      []uint32 // The transactions in pool
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
	State []TXNAttrInfo // the result from each validator
}

func GetLogEvent(obj *event.LogEventArgs) (map[string]bool, LogEventArgs) {
	hash := obj.TxHash
	addr := obj.ContractAddress.ToHexString()
	contractAddrs := map[string]bool{addr: true}
	return contractAddrs, LogEventArgs{hash.ToHexString(), addr, obj.Message}
}

func GetExecuteNotify(obj *event.ExecuteNotify) (map[string]bool, ExecuteNotify) {
	evts := []NotifyEventInfo{}
	var contractAddrs = make(map[string]bool)
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.ContractAddress.ToHexString(), v.States})
		contractAddrs[v.ContractAddress.ToHexString()] = true
	}
	txhash := obj.TxHash.ToHexString()
	return contractAddrs, ExecuteNotify{txhash, obj.State, obj.GasConsumed, evts}
}

func ConvertPreExecuteResult(obj *cstate.PreExecResult) PreExecuteResult {
	evts := []NotifyEventInfo{}
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.ContractAddress.ToHexString(), v.States})
	}
	return PreExecuteResult{obj.State, obj.Gas, obj.Result, evts}
}

func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.Nonce = ptx.Nonce
	trans.GasLimit = ptx.GasLimit
	trans.GasPrice = ptx.GasPrice
	trans.Payer = ptx.Payer.ToBase58()
	trans.Payload = TransPayloadToHex(ptx.Payload)

	trans.Attributes = make([]TxAttributeInfo, 0)
	trans.Sigs = []Sig{}
	for _, sigdata := range ptx.Sigs {
		sig, _ := sigdata.GetSig()
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
	trans.Hash = mhash.ToHexString()
	return trans
}

func TransferCrossChainMsg(msg *types.CrossChainMsg, pks []keypair.PublicKey) string {
	if msg == nil {
		return ""
	}
	sink := common.NewZeroCopySink(nil)
	msg.Serialization(sink)
	sink.WriteVarUint(uint64(len(pks)))
	for _, pk := range pks {
		key := keypair.SerializePublicKey(pk)
		sink.WriteVarBytes(key)
	}
	return common.ToHexString(sink.Bytes())
}

func SendTxToPool(txn *types.Transaction) (ontErrors.ErrCode, string) {
	if errCode, desc := bactor.AppendTxToPool(txn); errCode != ontErrors.ErrNoError {
		log.Warn("TxnPool verify error:", errCode.Error())
		return errCode, desc
	}
	return ontErrors.ErrNoError, ""
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
		PrevBlockHash:    block.Header.PrevBlockHash.ToHexString(),
		TransactionsRoot: block.Header.TransactionsRoot.ToHexString(),
		BlockRoot:        block.Header.BlockRoot.ToHexString(),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		ConsensusPayload: common.ToHexString(block.Header.ConsensusPayload),
		NextBookkeeper:   block.Header.NextBookkeeper.ToBase58(),
		Bookkeepers:      bookkeepers,
		SigData:          sigData,
		Hash:             hash.ToHexString(),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		tran := TransArryByteToHexString(block.Transactions[i])
		tran.Height = block.Header.Height
		trans[i] = tran
	}

	b := BlockInfo{
		Hash:         hash.ToHexString(),
		Size:         len(block.ToArray()),
		Header:       blockHead,
		Transactions: trans,
	}
	return b
}

func GetBalance(address common.Address) (*BalanceOfRsp, error) {
	balances, height, err := GetContractBalance(0, []common.Address{utils.OntContractAddress, utils.OngContractAddress}, address, true)
	if err != nil {
		return nil, fmt.Errorf("get ont balance error:%s", err)
	}
	return &BalanceOfRsp{
		Ont:    fmt.Sprintf("%d", balances[0]),
		Ong:    fmt.Sprintf("%d", balances[1]),
		Height: fmt.Sprintf("%d", height),
	}, nil
}

func GetGrantOng(addr common.Address) (string, error) {
	key := append([]byte(ont.UNBOUND_TIME_OFFSET), addr[:]...)
	value, err := ledger.DefLedger.GetStorageItem(utils.OntContractAddress, key)
	if err != nil {
		value = []byte{0, 0, 0, 0}
	}
	source := common.NewZeroCopySource(value)
	v, eof := source.NextUint32()
	if eof {
		return fmt.Sprintf("%v", 0), io.ErrUnexpectedEOF
	}
	onts, _, err := GetContractBalance(0, []common.Address{utils.OntContractAddress}, addr, false)
	if err != nil {
		return fmt.Sprintf("%v", 0), err
	}
	boundong := utils.CalcUnbindOng(onts[0], v, uint32(time.Now().Unix())-constants.GENESIS_BLOCK_TIMESTAMP)
	return fmt.Sprintf("%v", boundong), nil
}

func GetAllowance(asset string, from, to common.Address) (string, error) {
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case "ont":
		contractAddr = utils.OntContractAddress
	case "ong":
		contractAddr = utils.OngContractAddress
	default:
		return "", fmt.Errorf("unsupport asset")
	}
	allowance, err := GetContractAllowance(0, contractAddr, from, to)
	if err != nil {
		return "", fmt.Errorf("get allowance error:%s", err)
	}
	return fmt.Sprintf("%v", allowance), nil
}

func GetContractBalance(cVersion byte, contractAddres []common.Address, accAddr common.Address, atomic bool) ([]uint64, uint32, error) {
	txes := make([]*types.Transaction, 0, len(contractAddres))
	for _, contractAddr := range contractAddres {
		mutable, err := NewNativeInvokeTransaction(0, 0, contractAddr, cVersion, "balanceOf", []interface{}{accAddr[:]})
		if err != nil {
			return nil, 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
		}
		tx, err := mutable.IntoImmutable()
		if err != nil {
			return nil, 0, err
		}

		txes = append(txes, tx)
	}

	results, height, err := bactor.PreExecuteContractBatch(txes, atomic)
	if err != nil {
		return nil, 0, fmt.Errorf("PrepareInvokeContract error:%s", err)
	}
	balances := make([]uint64, 0, len(contractAddres))
	for _, result := range results {
		if result.State == 0 {
			return nil, 0, fmt.Errorf("prepare invoke failed")
		}
		data, err := hex.DecodeString(result.Result.(string))
		if err != nil {
			return nil, 0, fmt.Errorf("hex.DecodeString error:%s", err)
		}

		balance := common.BigIntFromNeoBytes(data)
		balances = append(balances, balance.Uint64())
	}

	return balances, height, nil
}

func GetContractAllowance(cVersion byte, contractAddr, fromAddr, toAddr common.Address) (uint64, error) {
	type allowanceStruct struct {
		From common.Address
		To   common.Address
	}
	mutable, err := NewNativeInvokeTransaction(0, 0, contractAddr, cVersion, "allowance",
		[]interface{}{&allowanceStruct{
			From: fromAddr,
			To:   toAddr,
		}})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}

	result, err := bactor.PreExecuteContract(tx)
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
	allowance := common.BigIntFromNeoBytes(data)
	return allowance.Uint64(), nil
}

func GetGasPrice() (map[string]interface{}, error) {
	start := bactor.GetCurrentBlockHeight()
	var gasPrice uint64 = 0
	var height uint32 = 0
	var end uint32 = 0
	if start > MAX_SEARCH_HEIGHT {
		end = start - MAX_SEARCH_HEIGHT
	}
	for i := start; i >= end; i-- {
		head, err := bactor.GetHeaderByHeight(i)
		if err == nil && head.TransactionsRoot != common.UINT256_EMPTY {
			height = i
			blk, err := bactor.GetBlockByHeight(i)
			if err != nil {
				return nil, err
			}
			for _, v := range blk.Transactions {
				gasPrice += v.GasPrice
			}
			gasPrice = gasPrice / uint64(len(blk.Transactions))
			break
		}
	}
	result := map[string]interface{}{"gasprice": gasPrice, "height": height}
	return result, nil
}

func GetBlockTransactions(block *types.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		t := block.Transactions[i].Hash()
		trans[i] = t.ToHexString()
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         hash.ToHexString(),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
}

//NewNativeInvokeTransaction return native contract invoke transaction
func NewNativeInvokeTransaction(gasPirce, gasLimit uint64, contractAddress common.Address, version byte,
	method string, params []interface{}) (*types.MutableTransaction, error) {
	invokeCode, err := cutils.BuildNativeInvokeCode(contractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	return NewSmartContractTransaction(gasPirce, gasLimit, invokeCode)
}

func NewNeovmInvokeTransaction(gasPrice, gasLimit uint64, contractAddress common.Address, params []interface{}) (*types.MutableTransaction, error) {
	invokeCode, err := cutils.BuildNeoVMInvokeCode(contractAddress, params)
	if err != nil {
		return nil, err
	}
	return NewSmartContractTransaction(gasPrice, gasLimit, invokeCode)
}

func NewSmartContractTransaction(gasPrice, gasLimit uint64, invokeCode []byte) (*types.MutableTransaction, error) {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.InvokeNeo,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     nil,
	}
	return tx, nil
}

//BuildNeoVMInvokeCode build NeoVM Invoke code for params
func BuildNeoVMInvokeCode(smartContractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := cutils.BuildNeoVMParam(builder, params)
	if err != nil {
		return nil, err
	}
	args := append(builder.ToArray(), 0x67)
	args = append(args, smartContractAddress[:]...)
	return args, nil
}

func GetAddress(str string) (common.Address, error) {
	var address common.Address
	var err error
	if len(str) == common.ADDR_LEN*2 {
		address, err = common.AddressFromHexString(str)
	} else {
		address, err = common.AddressFromBase58(str)
	}
	return address, err
}

type SyncStatus struct {
	CurrentBlockHeight uint32
	ConnectCount       uint32
	MaxPeerBlockHeight uint64
}

func GetSyncStatus() (SyncStatus, error) {
	height := bactor.GetMaxPeerBlockHeight()
	cnt := bactor.GetConnectionCnt()
	curBlockHeight := bactor.GetCurrentBlockHeight()

	return SyncStatus{
		CurrentBlockHeight: curBlockHeight,
		ConnectCount:       cnt,
		MaxPeerBlockHeight: height,
	}, nil
}
