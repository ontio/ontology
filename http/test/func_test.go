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

package test

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/vm/neovm"
	vmtypes "github.com/ontio/ontology/vm/types"
)

func TestCodeHash(t *testing.T) {
	code, _ := common.HexToBytes("aa")
	vmcode := vmtypes.VmCode{vmtypes.NEOVM, code}
	codehash := vmcode.AddressFromVmCode()
	fmt.Println(codehash.ToHexString())
	os.Exit(0)
}

func TestTxDeserialize(t *testing.T) {
	bys, _ := common.HexToBytes("")
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		fmt.Print("Deserialize Err:", err)
		os.Exit(0)
	}
	fmt.Printf("TxType:%x\n", txn.TxType)
	os.Exit(0)
}
func TestAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	pk, err := keypair.DeserializePublicKey(pubkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	ui60 := types.AddressFromPubKey(pk)
	addr := common.ToHexString(ui60[:])
	fmt.Println(addr)
	fmt.Println(ui60.ToBase58())
}
func TestMultiPubKeysAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	pk, err := keypair.DeserializePublicKey(pubkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	pubkey2, _ := common.HexToBytes("12020225c98cc5f82506fb9d01bad15a7be3da29c97a279bb6b55da1a3177483ab149b")
	pk2, err := keypair.DeserializePublicKey(pubkey2)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	ui60, _ := types.AddressFromMultiPubKeys([]*keypair.PublicKey{pk, pk2}, 1)
	addr := common.ToHexString(ui60[:])
	fmt.Println(addr)
	fmt.Println(ui60.ToBase58())
}

func TestInvokefunction(t *testing.T) {
	var funcName string
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := BuildSmartContractParamInter(builder, []interface{}{funcName, "", ""})
	if err != nil {
	}
	codeParams := builder.ToArray()
	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   codeParams,
	})
	tx.Nonce = uint32(time.Now().Unix())

	acct := account.Open(account.WALLET_FILENAME, []byte("passwordtest"))
	acc, err := acct.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}
	hash := tx.Hash()
	sign, _ := signature.Sign(acc.PrivateKey, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []keypair.PublicKey{acc.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	common.ToHexString(txbf.Bytes())
}
func BuildSmartContractParamInter(builder *neovm.ParamsBuilder, smartContractParams []interface{}) error {
	//虚拟机参数入栈时会反序
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
			err := BuildSmartContractParamInter(builder, v)
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
