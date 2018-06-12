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

package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	sig "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	httpcom "github.com/ontio/ontology/http/base/common"
	rpccommon "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	cstates "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION_TRANSACTION    = byte(0)
	VERSION_CONTRACT_ONT   = byte(0)
	VERSION_CONTRACT_ONG   = byte(0)
	CONTRACT_TRANSFER      = "transfer"
	CONTRACT_TRANSFER_FROM = "transferFrom"
	CONTRACT_APPROVE       = "approve"

	ASSET_ONT = "ont"
	ASSET_ONG = "ong"
)

//Return balance of address in base58 code
func GetBalance(address string) (*httpcom.BalanceOfRsp, error) {
	result, err := sendRpcRequest("getbalance", []interface{}{address})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	balance := &httpcom.BalanceOfRsp{}
	err = json.Unmarshal(result, balance)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

func GetAllowance(asset, from, to string) (string, error) {
	result, err := sendRpcRequest("getallowance", []interface{}{asset, from, to})
	if err != nil {
		return "", fmt.Errorf("sendRpcRequest error:%s", err)
	}
	balance := ""
	err = json.Unmarshal(result, &balance)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

//Transfer ont|ong from account to another account
func Transfer(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	transferTx, err := TransferTx(gasPrice, gasLimit, asset, signer.Address.ToBase58(), to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, transferTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(transferTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func TransferFrom(gasPrice, gasLimit uint64, signer *account.Account, asset, sender, from, to string, amount uint64) (string, error) {
	transferFromTx, err := TransferFromTx(gasPrice, gasLimit, asset, sender, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, transferFromTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(transferFromTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func Approve(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	approveTx, err := ApproveTx(gasPrice, gasLimit, asset, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, approveTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(approveTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func ApproveTx(gasPrice, gasLimit uint64, asset string, from, to string, amount uint64) (*types.Transaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	var state = &ont.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	}
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_ONT:
		version = VERSION_CONTRACT_ONT
		contractAddr = utils.OntContractAddress
	case ASSET_ONG:
		version = VERSION_CONTRACT_ONG
		contractAddr = utils.OngContractAddress
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	invokeCode, err := httpcom.BuildNativeInvokeCode(contractAddr, version, CONTRACT_APPROVE, []interface{}{state})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx, nil
}

func TransferTx(gasPrice, gasLimit uint64, asset, from, to string, amount uint64) (*types.Transaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	})
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_ONT:
		version = VERSION_CONTRACT_ONT
		contractAddr = utils.OntContractAddress
	case ASSET_ONG:
		version = VERSION_CONTRACT_ONG
		contractAddr = utils.OngContractAddress
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	invokeCode, err := httpcom.BuildNativeInvokeCode(contractAddr, version, CONTRACT_TRANSFER, []interface{}{sts})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx, nil
}

func TransferFromTx(gasPrice, gasLimit uint64, asset, sender, from, to string, amount uint64) (*types.Transaction, error) {
	senderAddr, err := common.AddressFromBase58(sender)
	if err != nil {
		return nil, fmt.Errorf("sender address:%s invalid:%s", to, err)
	}
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	transferFrom := &ont.TransferFrom{
		Sender: senderAddr,
		From:   fromAddr,
		To:     toAddr,
		Value:  amount,
	}
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_ONT:
		version = VERSION_CONTRACT_ONT
		contractAddr = utils.OntContractAddress
	case ASSET_ONG:
		version = VERSION_CONTRACT_ONG
		contractAddr = utils.OngContractAddress
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	invokeCode, err := httpcom.BuildNativeInvokeCode(contractAddr, version, CONTRACT_TRANSFER_FROM, []interface{}{transferFrom})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx, nil
}

func SignTransaction(signer *account.Account, tx *types.Transaction) error {
	tx.Payer = signer.Address
	txHash := tx.Hash()
	sigData, err := Sign(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}
	sig := &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sigData},
	}
	tx.Sigs = []*types.Sig{sig}
	return nil
}

//Sign sign return the signature to the data of private key
func Sign(data []byte, signer *account.Account) ([]byte, error) {
	s, err := sig.Sign(signer.SigScheme, signer.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := sig.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("sig.Serialize error:%s", err)
	}
	return sigData, nil
}

//SendRawTransaction send a transaction to ontology network, and return hash of the transaction
func SendRawTransaction(tx *types.Transaction) (string, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return "", fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData})
	if err != nil {
		return "", err
	}
	hexHash := ""
	err = json.Unmarshal(data, &hexHash)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal hash:%s error:%s", data, err)
	}
	return hexHash, nil
}

//GetSmartContractEvent return smart contract event execute by invoke transaction by hex string code
func GetSmartContractEvent(txHash string) (*rpccommon.ExecuteNotify, error) {
	data, err := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	notifies := &rpccommon.ExecuteNotify{}
	err = json.Unmarshal(data, &notifies)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal SmartContactEvent:%s error:%s", data, err)
	}
	return notifies, nil
}

func GetSmartContractEventInfo(txHash string) ([]byte, error) {
	return sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
}

func GetRawTransaction(txHash string) ([]byte, error) {
	return sendRpcRequest("getrawtransaction", []interface{}{txHash, 1})
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {
	return sendRpcRequest("getblock", []interface{}{hashOrHeight, 1})
}

func GetBlockData(hashOrHeight interface{}) ([]byte, error) {
	data, err := sendRpcRequest("getblock", []interface{}{hashOrHeight})
	if err != nil {
		return nil, err
	}
	hexStr := ""
	err = json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	blockData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return blockData, nil
}

func GetBlockCount() (uint32, error) {
	data, err := sendRpcRequest("getblockcount", []interface{}{})
	if err != nil {
		return 0, err
	}
	num := uint32(0)
	err = json.Unmarshal(data, &num)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return num, nil
}

func DeployContract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	needStorage bool,
	code,
	cname,
	cversion,
	cauthor,
	cemail,
	cdesc string) (string, error) {

	c, err := hex.DecodeString(code)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString error:%s", err)
	}
	tx := NewDeployCodeTransaction(gasPrice, gasLimit, c, needStorage, cname, cversion, cauthor, cemail, cdesc)

	err = SignTransaction(signer, tx)
	if err != nil {
		return "", err
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendRawTransaction error:%s", err)
	}
	return txHash, nil
}

func InvokeNativeContract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	contractAddress common.Address,
	version byte,
	method string,
	params []interface{},
) (string, error) {
	tx, err := httpcom.NewNativeInvokeTransaction(gasPrice, gasLimit, contractAddress, version, method, params)
	if err != nil {
		return "", err
	}
	return InvokeSmartContract(signer, tx)
}

//Invoke wasm smart contract
//methodName is wasm contract action name
//paramType  is Json or Raw format
//version should be greater than 0 (0 is reserved for test)
func InvokeWasmVMContract(
	gasPrice,
	gasLimit uint64,
	siger *account.Account,
	cversion byte, //version of contract
	contractAddress common.Address,
	method string,
	paramType wasmvm.ParamType,
	params []interface{}) (string, error) {

	invokeCode, err := BuildWasmVMInvokeCode(contractAddress, method, paramType, cversion, params)
	if err != nil {
		return "", err
	}
	tx, err := httpcom.NewSmartContractTransaction(gasPrice, gasLimit, invokeCode)
	if err != nil {
		return "", err
	}
	return InvokeSmartContract(siger, tx)
}

//Invoke neo vm smart contract. if isPreExec is true, the invoke will not really execute
func InvokeNeoVMContract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {
	tx, err := httpcom.NewNeovmInvokeTransaction(gasPrice, gasLimit, smartcodeAddress, params)
	if err != nil {
		return "", err
	}
	return InvokeSmartContract(signer, tx)
}

//InvokeSmartContract is low level method to invoke contact.
func InvokeSmartContract(signer *account.Account, tx *types.Transaction) (string, error) {
	err := SignTransaction(signer, tx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func PrepareInvokeNeoVMContract(
	contractAddress common.Address,
	params []interface{},
) (*cstates.PreExecResult, error) {
	tx, err := httpcom.NewNeovmInvokeTransaction(0, 0, contractAddress, params)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func PrepareInvokeCodeNeoVMContract(code []byte) (*cstates.PreExecResult, error) {
	tx, err := httpcom.NewSmartContractTransaction(0, 0, code)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func PrepareInvokeNativeContract(
	contractAddress common.Address,
	version byte,
	method string,
	params []interface{}) (*cstates.PreExecResult, error) {
	tx, err := httpcom.NewNativeInvokeTransaction(0, 0, contractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

//NewDeployCodeTransaction return a smart contract deploy transaction instance
func NewDeployCodeTransaction(gasPrice, gasLimit uint64, code []byte, needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {

	deployPayload := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        cname,
		Version:     cversion,
		Author:      cauthor,
		Email:       cemail,
		Description: cdesc,
	}
	tx := &types.Transaction{
		Version:  VERSION_TRANSACTION,
		TxType:   types.Deploy,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  deployPayload,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx
}

//for wasm vm
//build param bytes for wasm contract
func buildWasmContractParam(params []interface{}, paramType wasmvm.ParamType) ([]byte, error) {
	switch paramType {
	case wasmvm.Json:
		args := make([]exec.Param, len(params))

		for i, param := range params {
			switch param.(type) {
			case string:
				arg := exec.Param{Ptype: "string", Pval: param.(string)}
				args[i] = arg
			case int:
				arg := exec.Param{Ptype: "int", Pval: strconv.Itoa(param.(int))}
				args[i] = arg
			case int64:
				arg := exec.Param{Ptype: "int64", Pval: strconv.FormatInt(param.(int64), 10)}
				args[i] = arg
			case []int:
				bf := bytes.NewBuffer(nil)
				array := param.([]int)
				for i, tmp := range array {
					bf.WriteString(strconv.Itoa(tmp))
					if i != len(array)-1 {
						bf.WriteString(",")
					}
				}
				arg := exec.Param{Ptype: "int_array", Pval: bf.String()}
				args[i] = arg
			case []int64:
				bf := bytes.NewBuffer(nil)
				array := param.([]int64)
				for i, tmp := range array {
					bf.WriteString(strconv.FormatInt(tmp, 10))
					if i != len(array)-1 {
						bf.WriteString(",")
					}
				}
				arg := exec.Param{Ptype: "int_array", Pval: bf.String()}
				args[i] = arg
			default:
				return nil, fmt.Errorf("not a supported type :%v\n", param)
			}
		}

		bs, err := json.Marshal(exec.Args{args})
		if err != nil {
			return nil, err
		}
		return bs, nil
	case wasmvm.Raw:
		bf := bytes.NewBuffer(nil)
		for _, param := range params {
			switch param.(type) {
			case string:
				tmp := bytes.NewBuffer(nil)
				serialization.WriteString(tmp, param.(string))
				bf.Write(tmp.Bytes())

			case int:
				tmpBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(tmpBytes, uint32(param.(int)))
				bf.Write(tmpBytes)

			case int64:
				tmpBytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(tmpBytes, uint64(param.(int64)))
				bf.Write(tmpBytes)

			default:
				return nil, fmt.Errorf("not a supported type :%v\n", param)
			}
		}
		return bf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

//BuildWasmVMInvokeCode return wasn vm invoke code
func BuildWasmVMInvokeCode(smartcodeAddress common.Address, methodName string, paramType wasmvm.ParamType, version byte, params []interface{}) ([]byte, error) {
	contract := &cstates.Contract{}
	contract.Address = smartcodeAddress
	contract.Method = methodName
	contract.Version = version

	argbytes, err := buildWasmContractParam(params, paramType)

	if err != nil {
		return nil, fmt.Errorf("build wasm contract param failed:%s", err)
	}
	contract.Args = argbytes
	bf := bytes.NewBuffer(nil)
	contract.Serialize(bf)
	return bf.Bytes(), nil
}

//ParseNeoVMContractReturnTypeBool return bool value of smart contract execute code.
func ParseNeoVMContractReturnTypeBool(hexStr string) (bool, error) {
	return hexStr == "01", nil
}

//ParseNeoVMContractReturnTypeInteger return integer value of smart contract execute code.
func ParseNeoVMContractReturnTypeInteger(hexStr string) (int64, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return common.BigIntFromNeoBytes(data).Int64(), nil
}

//ParseNeoVMContractReturnTypeByteArray return []byte value of smart contract execute code.
func ParseNeoVMContractReturnTypeByteArray(hexStr string) (string, error) {
	return hexStr, nil
}

//ParseNeoVMContractReturnTypeString return string value of smart contract execute code.
func ParseNeoVMContractReturnTypeString(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString:%s error:%s", hexStr, err)
	}
	return string(data), nil
}
