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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/core/xshard_types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	cstate "github.com/ontio/ontology/smartcontract/states"
	"strings"
	"time"
)

const MAX_SEARCH_HEIGHT uint32 = 100
const MAX_REQUEST_BODY_SIZE = 1 << 20

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
	SourceTxHash    string
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

type CrossShardTransactions struct {
	ShardMsg *types.CrossShardMsgInfo `json:"shard_msg"`
	Tx       *Transactions
}

type Transactions struct {
	Version    byte
	ShardId    uint64
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
	ShardTxs     map[uint64][]*CrossShardTransactions
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
		sourceTxHash := ""
		if v.SourceTxHash != common.UINT256_EMPTY {
			sourceTxHash = v.SourceTxHash.ToHexString()
		}
		evts = append(evts, NotifyEventInfo{v.ContractAddress.ToHexString(), v.States, sourceTxHash})
		contractAddrs[v.ContractAddress.ToHexString()] = true
	}
	txhash := obj.TxHash.ToHexString()
	return contractAddrs, ExecuteNotify{txhash, obj.State, obj.GasConsumed, evts}
}

func ConvertPreExecuteResult(obj *cstate.PreExecResult) PreExecuteResult {
	evts := []NotifyEventInfo{}
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.ContractAddress.ToHexString(), v.States, v.SourceTxHash.ToHexString()})
	}
	return PreExecuteResult{obj.State, obj.Gas, obj.Result, evts}
}

func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.ShardId = ptx.ShardID.ToUint64()
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
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	shardTxs := make(map[uint64][]*CrossShardTransactions)
	for shardId, infos := range block.ShardTxs {
		trans := make([]*CrossShardTransactions, 0)
		for _, info := range infos {
			trans = append(trans, &CrossShardTransactions{
				ShardMsg: info.ShardMsg, Tx: TransArryByteToHexString(info.Tx),
			})
		}
		shardTxs[shardId.ToUint64()] = trans
	}

	b := BlockInfo{
		Hash:         hash.ToHexString(),
		Size:         len(block.ToArray()),
		ShardTxs:     shardTxs,
		Header:       blockHead,
		Transactions: trans,
	}
	return b
}

func GetBalance(address common.Address) (*BalanceOfRsp, error) {
	ont, err := GetContractBalance(0, utils.OntContractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get ont balance error:%s", err)
	}
	ong, err := GetContractBalance(0, utils.OngContractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get ont balance error:%s", err)
	}
	return &BalanceOfRsp{
		Ont: fmt.Sprintf("%d", ont),
		Ong: fmt.Sprintf("%d", ong),
	}, nil
}

func GetGrantOng(addr common.Address) (string, error) {
	key := append([]byte(ont.UNBOUND_TIME_OFFSET), addr[:]...)
	value, err := ledger.DefLedger.GetStorageItem(utils.OntContractAddress, key)
	if err != nil {
		value = []byte{0, 0, 0, 0}
	}
	v, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return fmt.Sprintf("%v", 0), err
	}
	ont, err := GetContractBalance(0, utils.OntContractAddress, addr)
	if err != nil {
		return fmt.Sprintf("%v", 0), err
	}
	boundong := utils.CalcUnbindOng(ont, v, uint32(time.Now().Unix())-constants.GENESIS_BLOCK_TIMESTAMP)
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

func GetContractBalance(cVersion byte, contractAddr, accAddr common.Address) (uint64, error) {
	mutable, err := NewNativeInvokeTransaction(0, 0, contractAddr, cVersion, "balanceOf", []interface{}{accAddr[:]})
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

	balance := common.BigIntFromNeoBytes(data)
	return balance.Uint64(), nil
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
		Version:  common.CURR_TX_VERSION,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     nil,
	}
	return tx, nil
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

type TxStateInfo struct {
	TxID           string           // cross shard tx id: userTxHash+notify1+notify2...
	Shards         map[uint64]uint8 // shards in this shard transaction, not include notification
	NumNotifies    uint32
	ShardNotifies  []XShardNotify
	NextReqID      uint32
	InReqResp      map[uint64][]XShardTxReqResp // todo: request id may conflict
	PendingInReq   XShardTxReq
	TotalInReq     uint32
	OutReqResp     []XShardTxReqResp
	TxPayload      string
	PendingOutReq  XShardTxReq
	PendingPrepare *xshard_types.XShardPrepareMsg
	ExecState      uint8
	Result         string
	ResultErr      string
	LockedAddress  []string
	LockedKeys     []string
	WriteSet       *overlaydb.MemDB
	Notify         ExecuteNotify
}
type ShardMsgHeader struct {
	ShardTxID     string
	SourceShardID uint64
	TargetShardID uint64
	SourceTxHash  string
}
type XShardTxReq struct {
	ShardMsgHeader
	IdxInTx  uint64
	Contract string
	Payer    string
	Fee      uint64
	GasPrice uint64
	Method   string
	Args     string
}
type XShardTxRsp struct {
	ShardMsgHeader
	IdxInTx uint64
	FeeUsed uint64
	Error   bool
	Result  string
}
type XShardTxReqResp struct {
	Req   XShardTxReq
	Resp  XShardTxRsp
	Index uint32
}

type XShardNotify struct {
	ShardMsgHeader
	NotifyID uint32
	Contract string
	Payer    string
	Fee      uint64
	Method   string
	Args     string
}

func ParseShardState(txState *xshard_state.TxState) (TxStateInfo, error) {

	shards := make(map[uint64]uint8)
	for k, v := range txState.Shards {
		shards[k.ToUint64()] = uint8(v)
	}
	lockedAddress := make([]string, 0)
	for _, addr := range txState.LockedAddress {
		lockedAddress = append(lockedAddress, addr.ToBase58())
	}
	lockedKeys := make([]string, 0)
	for _, key := range txState.LockedKeys {
		lockedKeys = append(lockedKeys, common.ToHexString(key))
	}
	var notify ExecuteNotify
	if txState.Notify != nil {
		_, notify = GetExecuteNotify(txState.Notify)
	}
	inReqResp := make(map[uint64][]XShardTxReqResp)
	if txState.InReqResp != nil {
		for k, v := range txState.InReqResp {
			xShardTxReqResps := parseXShardTxReqResp(v)
			inReqResp[k.ToUint64()] = xShardTxReqResps
		}
	}
	var pendingInReq XShardTxReq
	if txState.PendingInReq != nil {
		pendingInReq, _ = parseXShardTxReq(txState.PendingInReq)
	}
	var outReqResp []XShardTxReqResp
	if txState.OutReqResp != nil {
		outReqResp = parseXShardTxReqResp(txState.OutReqResp)
	}
	var pendingOutReq XShardTxReq
	if txState.PendingOutReq != nil {
		pendingOutReq, _ = parseXShardTxReq(txState.PendingOutReq)
	}
	var xns []XShardNotify
	if txState.ShardNotifies != nil {
		xns = parseShardNotifies(txState.ShardNotifies)
	}
	return TxStateInfo{
		TxID:           common.ToHexString([]byte(string(txState.TxID))),
		Shards:         shards,
		NumNotifies:    txState.NumNotifies,
		ShardNotifies:  xns,
		NextReqID:      txState.NextReqID,
		InReqResp:      inReqResp,
		PendingInReq:   pendingInReq,
		TotalInReq:     txState.TotalInReq,
		OutReqResp:     outReqResp,
		TxPayload:      common.ToHexString(txState.TxPayload),
		PendingOutReq:  pendingOutReq,
		PendingPrepare: txState.PendingPrepare,
		ExecState:      uint8(txState.ExecState),
		Result:         common.ToHexString(txState.Result),
		ResultErr:      txState.ResultErr,
		LockedAddress:  lockedAddress,
		LockedKeys:     lockedKeys,
		WriteSet:       txState.WriteSet,
		Notify:         notify,
	}, nil
}

func parseShardNotifies(notifys []*xshard_types.XShardNotify) []XShardNotify {
	xShardnotifys := make([]XShardNotify, 0, len(notifys))
	for _, xNotify := range notifys {
		header := ShardMsgHeader{
			ShardTxID:     common.ToHexString([]byte(string(xNotify.ShardMsgHeader.ShardTxID))),
			SourceShardID: xNotify.ShardMsgHeader.SourceShardID.ToUint64(),
			TargetShardID: xNotify.ShardMsgHeader.TargetShardID.ToUint64(),
			SourceTxHash:  xNotify.ShardMsgHeader.SourceTxHash.ToHexString(),
		}
		n := XShardNotify{
			ShardMsgHeader: header,
			NotifyID:       xNotify.NotifyID,
			Contract:       xNotify.Contract.ToHexString(),
			Payer:          xNotify.Payer.ToBase58(),
			Fee:            xNotify.Fee,
			Method:         xNotify.Method,
			Args:           common.ToHexString(xNotify.Args),
		}
		xShardnotifys = append(xShardnotifys, n)
	}
	return xShardnotifys
}

func parseXShardTxReqResp(reqRsp []*xshard_state.XShardTxReqResp) []XShardTxReqResp {
	xShardTxReqResps := make([]XShardTxReqResp, 0, len(reqRsp))
	for _, item := range reqRsp {
		xShardTxReq, shardMsgHeader := parseXShardTxReq(item.Req)
		xShardTxRsp := XShardTxRsp{
			ShardMsgHeader: shardMsgHeader,
			IdxInTx:        item.Resp.IdxInTx,
			FeeUsed:        item.Resp.FeeUsed,
			Error:          item.Resp.Error,
			Result:         common.ToHexString(item.Resp.Result),
		}
		xShardTxReqResp := XShardTxReqResp{
			Req:   xShardTxReq,
			Resp:  xShardTxRsp,
			Index: item.Index,
		}
		xShardTxReqResps = append(xShardTxReqResps, xShardTxReqResp)
	}
	return xShardTxReqResps
}
func parseXShardTxReq(req *xshard_types.XShardTxReq) (XShardTxReq, ShardMsgHeader) {

	shardMsgHeader := ShardMsgHeader{
		ShardTxID:     common.ToHexString([]byte(string(req.ShardTxID))),
		SourceShardID: req.SourceShardID.ToUint64(),
		TargetShardID: req.TargetShardID.ToUint64(),
		SourceTxHash:  req.SourceTxHash.ToHexString(),
	}
	xShardTxReq := XShardTxReq{
		ShardMsgHeader: shardMsgHeader,
		IdxInTx:        req.IdxInTx,
		Contract:       req.Contract.ToHexString(),
		Payer:          req.Payer.ToBase58(),
		Fee:            req.Fee,
		GasPrice:       req.GasPrice,
		Method:         req.Method,
		Args:           common.ToHexString(req.Args),
	}
	return xShardTxReq, shardMsgHeader
}
