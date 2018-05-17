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
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	rpccommon "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	cstates "github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/neovm"
	neotypes "github.com/ontio/ontology/vm/neovm/types"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION_TRANSACTION  = 0
	VERSION_CONTRACT_ONT = 0
	VERSION_CONTRACT_ONG = 0
	CONTRACT_TRANSFER    = "transfer"

	ASSET_ONT = "ont"
	ASSET_ONG = "ong"
)

//Return balance of address in base58 code
func GetBalance(add string) (*rpccommon.BalanceOfRsp, error) {
	data, err := sendRpcRequest("getbalance", []interface{}{add})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	rsp := &rpccommon.BalanceOfRsp{}
	err = json.Unmarshal(data, rsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return rsp, nil
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

func TransferTx(gasPrice, gasLimit uint64, asset, from, to string, amount uint64) (*types.Transaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	buf := bytes.NewBuffer(nil)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err = transfers.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("transfers.Serialize error %s", err)
	}
	var cversion byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_ONT:
		contractAddr = genesis.OntContractAddress
		cversion = VERSION_CONTRACT_ONT
	case ASSET_ONG:
		contractAddr = genesis.OngContractAddress
		cversion = VERSION_CONTRACT_ONG
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	return InvokeNativeContractTx(gasPrice, gasLimit, cversion, contractAddr, CONTRACT_TRANSFER, buf.Bytes())
}

func SignTransaction(signer *account.Account, tx *types.Transaction) error {
	tx.Payer = signer.Address
	txHash := tx.Hash()
	sigData, err := sign(signer.SigScheme.Name(), txHash.ToArray(), signer)
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
func sign(cryptScheme string, data []byte, signer *account.Account) ([]byte, error) {
	scheme, err := sig.GetScheme(cryptScheme)
	if err != nil {
		return nil, fmt.Errorf("GetScheme by:%s error:%s", cryptScheme, err)
	}
	s, err := sig.Sign(scheme, signer.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := sig.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("sig.Serialize error:%s", err)
	}
	return sigData, nil
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(gasPirce, gasLimit uint64, vmType vmtypes.VmType, code []byte) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: vmtypes.VmCode{
			VmType: vmType,
			Code:   code,
		},
	}
	tx := &types.Transaction{
		Version:    VERSION_TRANSACTION,
		GasPrice:   gasPirce,
		GasLimit:   gasLimit,
		TxType:     types.Invoke,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    invokePayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
		Sigs:       make([]*types.Sig, 0, 0),
	}
	return tx
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

func GetRawTransaction(txHash string) ([]byte, error) {
	data, err := sendRpcRequest("getrawtransaction", []interface{}{txHash, 1})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return data, nil
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {
	data, err := sendRpcRequest("getblock", []interface{}{hashOrHeight, 1})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DeployContract(
	gasPrice,
	gasLimit uint64,
	singer *account.Account,
	vmType vmtypes.VmType,
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
	tx := NewDeployCodeTransaction(gasPrice, gasLimit, vmType, c, needStorage, cname, cversion, cauthor, cemail, cdesc)

	err = SignTransaction(singer, tx)
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
	singer *account.Account,
	cversion byte,
	contractAddress common.Address,
	method string,
	args []byte,
) (string, error) {
	return InvokeSmartContract(gasPrice, gasLimit, singer, vmtypes.Native, cversion, contractAddress, method, args)
}

func InvokeNativeContractTx(gasPrice,
	gasLimit uint64,
	cversion byte,
	contractAddress common.Address,
	method string,
	args []byte) (*types.Transaction, error) {
	return InvokeSmartContractTx(gasPrice, gasLimit, vmtypes.Native, cversion, contractAddress, method, args)
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

	args, err := buildWasmContractParam(params, paramType)
	if err != nil {
		return "", fmt.Errorf("buildWasmContractParam error:%s", err)
	}
	return InvokeSmartContract(gasPrice, gasLimit, siger, vmtypes.WASMVM, cversion, contractAddress, method, args)
}

//Invoke neo vm smart contract. if isPreExec is true, the invoke will not really execute
func InvokeNeoVMContract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	cversion byte,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {

	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return "", err
	}
	args := builder.ToArray()

	return InvokeSmartContract(gasPrice, gasLimit, signer, vmtypes.NEOVM, cversion, smartcodeAddress, "", args)
}

func InvokeNeoVMContractTx(gasPrice,
	gasLimit uint64,
	cversion byte,
	smartcodeAddress common.Address,
	params []interface{}) (*types.Transaction, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return nil, err
	}
	args := builder.ToArray()
	return InvokeSmartContractTx(gasPrice, gasLimit, vmtypes.NEOVM, cversion, smartcodeAddress, "", args)
}

//InvokeSmartContract is low level method to invoke contact.
func InvokeSmartContract(
	gasPrice,
	gasLimit uint64,
	singer *account.Account,
	vmType vmtypes.VmType,
	cversion byte,
	contractAddress common.Address,
	method string,
	args []byte,
) (string, error) {
	invokeTx, err := InvokeSmartContractTx(gasPrice, gasLimit, vmType, cversion, contractAddress, method, args)
	if err != nil {
		return "", err
	}
	err = SignTransaction(singer, invokeTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(invokeTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func InvokeSmartContractTx(gasPrice,
	gasLimit uint64,
	vmType vmtypes.VmType,
	cversion byte,
	contractAddress common.Address,
	method string,
	args []byte) (*types.Transaction, error) {
	crt := &cstates.Contract{
		Version: cversion,
		Address: contractAddress,
		Method:  method,
		Args:    args,
	}
	buf := bytes.NewBuffer(nil)
	err := crt.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("Serialize contract error:%s", err)
	}
	invokCode := buf.Bytes()
	if vmType == vmtypes.NEOVM {
		invokCode = append([]byte{0x67}, invokCode[:]...)
	}
	return NewInvokeTransaction(gasPrice, gasLimit, vmType, invokCode), nil
}

func PrepareInvokeNeoVMContract(
	gasPrice,
	gasLimit uint64,
	cversion byte,
	contractAddress common.Address,
	params []interface{},
) (interface{}, error) {
	code, err := BuildNeoVMInvokeCode(cversion, contractAddress, params)
	if err != nil {
		return nil, fmt.Errorf("BuildNVMInvokeCode error:%s", err)
	}
	tx := NewInvokeTransaction(gasPrice, gasLimit, vmtypes.NEOVM, code)
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
	var res interface{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal hash:%s error:%s", data, err)
	}
	return res, nil
}

//NewDeployCodeTransaction return a smart contract deploy transaction instance
func NewDeployCodeTransaction(
	gasPrice,
	gasLimit uint64,
	vmType vmtypes.VmType,
	code []byte,
	needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {

	vmCode := vmtypes.VmCode{
		VmType: vmType,
		Code:   code,
	}
	deployPayload := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: needStorage,
		Name:        cname,
		Version:     cversion,
		Author:      cauthor,
		Email:       cemail,
		Description: cdesc,
	}
	tx := &types.Transaction{
		Version:    VERSION_TRANSACTION,
		TxType:     types.Deploy,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    deployPayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
		GasPrice:   gasPrice,
		GasLimit:   gasLimit,
		Sigs:       make([]*types.Sig, 0, 0),
	}
	return tx
}

//buildNeoVMParamInter build neovm invoke param code
func buildNeoVMParamInter(builder *neovm.ParamsBuilder, smartContractParams []interface{}) error {
	//VM load params in reverse order
	for i := len(smartContractParams) - 1; i >= 0; i-- {
		switch v := smartContractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case common.Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case []interface{}:
			err := buildNeoVMParamInter(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(neovm.PACK)
		default:
			return fmt.Errorf("unsupported param:%s", v)
		}
	}
	return nil
}

//BuildNeoVMInvokeCode build NeoVM Invoke code for params
func BuildNeoVMInvokeCode(cversion byte, smartContractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return nil, err
	}
	args := builder.ToArray()

	crt := &cstates.Contract{
		Version: cversion,
		Address: smartContractAddress,
		Args:    args,
	}
	crtBuf := bytes.NewBuffer(nil)
	err = crt.Serialize(crtBuf)
	if err != nil {
		return nil, fmt.Errorf("Serialize contract error:%s", err)
	}

	buf := bytes.NewBuffer(nil)
	buf.Write(append([]byte{0x67}, crtBuf.Bytes()[:]...))
	return buf.Bytes(), nil
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

//GetContractAddress return contract address
func GetContractAddress(code string, vmType vmtypes.VmType) common.Address {
	data, _ := hex.DecodeString(code)
	vmCode := &vmtypes.VmCode{
		VmType: vmType,
		Code:   data,
	}
	return vmCode.AddressFromVmCode()
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
	return neotypes.ConvertBytesToBigInteger(data).Int64(), nil
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
