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

package structPrinter

//import (
//	"bytes"
//	"fmt"
//	. "github.com/Ontology/common"
//	"github.com/Ontology/core/asset"
//	. "github.com/Ontology/core/contract"
//	tx "github.com/Ontology/core/transaction"
//	"github.com/Ontology/core/transaction/payload"
//	"github.com/Ontology/core/transaction/utxo"
//	"github.com/hokaccha/go-prettyjson"
//	"strconv"
//)
//
//type TxAttributeInfo struct {
//	Usage tx.TransactionAttributeUsage
//	Data  string
//}
//
//type UTXOTxInputInfo struct {
//	ReferTxID          string
//	ReferTxOutputIndex uint16
//}
//
//type BalanceTxInputInfo struct {
//	AssetID     string
//	Value       string
//	Address string
//}
//
//type TxoutputInfo struct {
//	AssetID     string
//	Value       string
//	Address string
//}
//
//type TxoutputMap struct {
//	Key   Uint256
//	Txout []TxoutputInfo
//}
//
//type AmountMap struct {
//	Key   Uint256
//	Value Fixed64
//}
//
//type ProgramInfo struct {
//	Code      string
//	Parameter string
//}
//
//type Transactions struct {
//	TxType         tx.TransactionType
//	PayloadVersion byte
//	Payload        PayloadInfo
//	Attributes     []TxAttributeInfo
//	UTXOInputs     []UTXOTxInputInfo
//	BalanceInputs  []BalanceTxInputInfo
//	Outputs        []TxoutputInfo
//	Programs       []ProgramInfo
//	NetworkFee     Fixed64
//	SystemFee      Fixed64
//
//	Hash string
//}
//
//type BlockHead struct {
//	Version          uint32
//	PrevBlockHash    string
//	TransactionsRoot string
//	Timestamp        uint32
//	Height           uint32
//	ConsensusData    uint64
//	NextBookKeeper   string
//	Program          ProgramInfo
//
//	Hash string
//}
//
//type BlockInfo struct {
//	Hash         string
//	BlockData    *BlockHead
//	Transactions []*Transactions
//}
//
//type TxInfo struct {
//	Hash string
//	Hex  string
//	Tx   *Transactions
//}
//
//type TxoutInfo struct {
//	High  uint32
//	Low   uint32
//	Txout utxo.TxOutput
//}
//
//type PayloadInfo interface{}
//
////implement PayloadInfo define BookKeepingInfo
//type BookKeepingInfo struct {
//	Nonce  uint64
//	Issuer IssuerInfo
//}
//
////implement PayloadInfo define DeployCodeInfo
//type FunctionCodeInfo struct {
//	Code           string
//	ParameterTypes string
//	ReturnTypes    string
//}
//
//type DeployCodeInfo struct {
//	Code        *FunctionCodeInfo
//	Name        string
//	CodeVersion string
//	Author      string
//	Email       string
//	Description string
//}
//
////implement PayloadInfo define IssueAssetInfo
//type IssueAssetInfo struct {
//}
//
//type IssuerInfo struct {
//	X, Y string
//}
//
////implement PayloadInfo define RegisterAssetInfo
//type RegisterAssetInfo struct {
//	Asset      *asset.Asset
//	Amount     Fixed64
//	Issuer     IssuerInfo
//	Controller string
//}
//
////implement PayloadInfo define TransferAssetInfo
//type TransferAssetInfo struct {
//}
//
//type RecordInfo struct {
//	RecordType string
//	RecordData string
//}
//
//type BookkeeperInfo struct {
//	PubKey     string
//	Action     string
//	Issuer     IssuerInfo
//	Controller string
//}
//
//type DataFileInfo struct {
//	IPFSPath string
//	Filename string
//	Note     string
//	Issuer   IssuerInfo
//}
//
//type Claim struct {
//	Claims []*UTXOTxInput
//}
//type UTXOTxInput struct {
//	ReferTxID          string
//	ReferTxOutputIndex uint16
//}
//
//type PrivacyPayloadInfo struct {
//	PayloadType uint8
//	Payload     string
//	EncryptType uint8
//	EncryptAttr string
//}
//
//func TransArryByteToHexString(ptx *tx.Transaction) {
//
//	trans := new(Transactions)
//	trans.TxType = ptx.TxType
//	trans.PayloadVersion = ptx.PayloadVersion
//	trans.Payload = TransPayloadToHex(ptx.Payload)
//
//	n := 0
//	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
//	for _, v := range ptx.Attributes {
//		trans.Attributes[n].Usage = v.Usage
//		trans.Attributes[n].Data = ToHexString(v.Data)
//		n++
//	}
//
//	n = 0
//	trans.UTXOInputs = make([]UTXOTxInputInfo, len(ptx.UTXOInputs))
//	for _, v := range ptx.UTXOInputs {
//		trans.UTXOInputs[n].ReferTxID = ToHexString(v.ReferTxID.ToArrayReverse())
//		trans.UTXOInputs[n].ReferTxOutputIndex = v.ReferTxOutputIndex
//		n++
//	}
//
//	n = 0
//	trans.BalanceInputs = make([]BalanceTxInputInfo, len(ptx.BalanceInputs))
//	for _, v := range ptx.BalanceInputs {
//		trans.BalanceInputs[n].AssetID = ToHexString(v.AssetID.ToArray())
//		trans.BalanceInputs[n].Value = strconv.FormatInt(int64(v.Value), 10)
//		trans.BalanceInputs[n].Address = ToHexString(v.Address.ToArray())
//		n++
//	}
//
//	n = 0
//	trans.Outputs = make([]TxoutputInfo, len(ptx.Outputs))
//	for _, v := range ptx.Outputs {
//		trans.Outputs[n].AssetID = ToHexString(v.AssetID.ToArrayReverse())
//		trans.Outputs[n].Value = strconv.FormatInt(int64(v.Value), 10)
//		trans.Outputs[n].Address = ToHexString(v.Address.ToArray())
//		n++
//	}
//
//	n = 0
//	trans.Programs = make([]ProgramInfo, len(ptx.Programs))
//	for _, v := range ptx.Programs {
//		trans.Programs[n].Code = ToHexString(v.Code)
//		trans.Programs[n].Parameter = ToHexString(v.Parameter)
//		n++
//	}
//
//	n = 0
//	trans.NetworkFee = ptx.NetworkFee
//	trans.SystemFee = ptx.SystemFee
//
//	mhash := ptx.Hash()
//	trans.Hash = ToHexString(mhash.ToArray())
//
//	str, _ := prettyjson.Marshal(trans)
//	fmt.Print(string(str))
//}
//
//func TransPayloadToHex(p tx.Payload) PayloadInfo {
//	switch object := p.(type) {
//	case *payload.BookKeeping:
//		obj := new(BookKeepingInfo)
//		obj.Nonce = object.Nonce
//		return obj
//	case *payload.BookKeeper:
//		obj := new(BookkeeperInfo)
//		encodedPubKey, _ := object.PubKey.EncodePoint(true)
//		obj.PubKey = ToHexString(encodedPubKey)
//		if object.Action == payload.BookKeeperAction_ADD {
//			obj.Action = "add"
//		} else if object.Action == payload.BookKeeperAction_SUB {
//			obj.Action = "sub"
//		} else {
//			obj.Action = "nil"
//		}
//		obj.Issuer.X = object.Issuer.X.String()
//		obj.Issuer.Y = object.Issuer.Y.String()
//
//		return obj
//	case *payload.IssueAsset:
//		obj := new(IssueAssetInfo)
//		return obj
//	case *payload.TransferAsset:
//		obj := new(TransferAssetInfo)
//		return obj
//	case *payload.DeployCode:
//		obj := new(DeployCodeInfo)
//		obj.Code.Code = ToHexString(object.Code.Code)
//		obj.Code.ParameterTypes = ToHexString(ContractParameterTypeToByte(object.Code.ParameterTypes))
//		obj.Code.ReturnTypes = ToHexString(ContractParameterTypeToByte(object.Code.ReturnTypes))
//		obj.Name = object.Name
//		obj.CodeVersion = object.CodeVersion
//		obj.Author = object.Author
//		obj.Email = object.Email
//		obj.Description = object.Description
//		return obj
//	case *payload.RegisterAsset:
//		obj := new(RegisterAssetInfo)
//		obj.Asset = object.Asset
//		obj.Amount = object.Amount
//		obj.Issuer.X = object.Issuer.X.String()
//		obj.Issuer.Y = object.Issuer.Y.String()
//		obj.Controller = ToHexString(object.Controller.ToArray())
//		return obj
//	case *payload.Record:
//		obj := new(RecordInfo)
//		obj.RecordType = object.RecordType
//		obj.RecordData = ToHexString(object.RecordData)
//		return obj
//	case *payload.PrivacyPayload:
//		obj := new(PrivacyPayloadInfo)
//		obj.PayloadType = uint8(object.PayloadType)
//		obj.Payload = ToHexString(object.Payload)
//		obj.EncryptType = uint8(object.EncryptType)
//		bytesBuffer := bytes.NewBuffer([]byte{})
//		object.EncryptAttr.Serialize(bytesBuffer)
//		obj.EncryptAttr = ToHexString(bytesBuffer.Bytes())
//		return obj
//	case *payload.DataFile:
//		obj := new(DataFileInfo)
//		obj.IPFSPath = object.IPFSPath
//		obj.Filename = object.Filename
//		obj.Note = object.Note
//		obj.Issuer.X = object.Issuer.X.String()
//		obj.Issuer.Y = object.Issuer.Y.String()
//		return obj
//	case *payload.Claim:
//		obj := new(Claim)
//		for _, v := range object.Claims {
//			item := new(UTXOTxInput)
//			item.ReferTxID = ToHexString(v.ReferTxID.ToArray())
//			item.ReferTxOutputIndex = v.ReferTxOutputIndex
//			obj.Claims = append(obj.Claims, item)
//		}
//		return obj
//	}
//	return nil
//}
