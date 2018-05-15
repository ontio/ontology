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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	sig "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	rpccommon "github.com/ontio/ontology/http/base/common"
	cstates "github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"math/big"
	"strconv"
	"time"

	"encoding/binary"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/vm/wasmvm/exec"
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

//Transfer ont from account to another account
func Transfer(signer *account.Account, to string, amount uint) (string, error) {
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return "", fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	buf := bytes.NewBuffer(nil)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  signer.Address,
		To:    toAddr,
		Value: uint64(amount),
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err = transfers.Serialize(buf)
	if err != nil {
		return "", fmt.Errorf("transfers.Serialize error %s", err)
	}
	crt := &cstates.Contract{
		Address: genesis.OntContractAddress,
		Method:  "transfer",
		Args:    buf.Bytes(),
	}
	buf = bytes.NewBuffer(nil)
	err = crt.Serialize(buf)
	if err != nil {
		return "", fmt.Errorf("Serialize contract error:%s", err)
	}
	invokeTx := NewInvokeTransaction(new(big.Int).SetInt64(0), vmtypes.Native, buf.Bytes())
	err = SignTransaction(signer, invokeTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	return SendRawTransaction(invokeTx)
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
func NewInvokeTransaction(gasLimit *big.Int, vmType vmtypes.VmType, code []byte) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: vmtypes.VmCode{
			VmType: vmType,
			Code:   code,
		},
	}
	tx := &types.Transaction{
		Version:    0,
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
func GetSmartContractEvent(txHash string) ([]*rpccommon.NotifyEventInfo, error) {
	data, err := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	events := make([]*rpccommon.NotifyEventInfo, 0)
	err = json.Unmarshal(data, &events)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal SmartContactEvent:%s error:%s", data, err)
	}
	return events, nil
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
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return data, nil
}

func DeployContract(singer *account.Account,
	vmType vmtypes.VmType,
	needStorage bool,
	code,
	name,
	version,
	author,
	email,
	desc string) (string, error) {

	c, err := hex.DecodeString(code)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString error:%s", err)
	}
	tx := NewDeployCodeTransaction(vmType, c, needStorage, name, version, author, email, desc)

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

//Invoke wasm smart contract
//methodName is wasm contract action name
//paramType  is Json or Raw format
//version should be greater than 0 (0 is reserved for test)
func InvokeWasmVMSmartContract(
	siger *account.Account,
	gasLimit *big.Int,
	smartcodeAddress common.Address,
	methodName string,
	paramType wasmvm.ParamType,
	version byte,
	params []interface{}) (string, error) {

	code, err := BuildWasmVMInvokeCode(smartcodeAddress, methodName, paramType, version, params)
	if err != nil {
		return "", err
	}
	tx := NewInvokeTransaction(new(big.Int), vmtypes.WASMVM, code)
	err = SignTransaction(siger, tx)
	if err != nil {
		return "", nil
	}
	return SendRawTransaction(tx)
}

//Invoke neo vm smart contract. if isPreExec is true, the invoke will not really execute
func InvokeNeoVMSmartContract(
	siger *account.Account,
	gasLimit *big.Int,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {
	code, err := BuildNeoVMInvokeCode(smartcodeAddress, params)
	if err != nil {
		return "", fmt.Errorf("BuildNVMInvokeCode error:%s", err)
	}
	tx := NewInvokeTransaction(gasLimit, vmtypes.NEOVM, code)
	err = SignTransaction(siger, tx)
	if err != nil {
		return "", nil
	}
	return SendRawTransaction(tx)
}

//NewDeployCodeTransaction return a smart contract deploy transaction instance
func NewDeployCodeTransaction(
	vmType vmtypes.VmType,
	code []byte,
	needStorage bool,
	name, version, author, email, desc string) *types.Transaction {

	vmCode := vmtypes.VmCode{
		VmType: vmType,
		Code:   code,
	}
	deployPayload := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: needStorage,
		Name:        name,
		Version:     version,
		Author:      author,
		Email:       email,
		Description: desc,
	}
	tx := &types.Transaction{
		Version:    0,
		TxType:     types.Deploy,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    deployPayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
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
func BuildNeoVMInvokeCode(smartContractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return nil, err
	}
	args := builder.ToArray()

	crt := &cstates.Contract{
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
