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
	"testing"
	"time"

	"encoding/json"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/merkle"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/stretchr/testify/assert"
)

func TestMerkleVerifier(t *testing.T) {
	type merkleProof struct {
		Type             string
		TransactionsRoot string
		BlockHeight      uint32
		CurBlockRoot     string
		CurBlockHeight   uint32
		TargetHashes     []string
	}
	proof := merkleProof{
		Type:             "MerkleProof",
		TransactionsRoot: "4b74e15973ce3964ba4a33ddaf92efbff922ea2225bca7676f62eab05829f11f",
		BlockHeight:      2,
		CurBlockRoot:     "a5094c1daeeceab46319ce62b600c68a7accc806bd9fe2fdb869560bf66b5251",
		CurBlockHeight:   6,
		TargetHashes: []string{
			"c7ac8087b4ce292d654001b1ab1bfe5e68fa6f7b8492a5b2f83560f8ac28f5fa",
			"5205a22b07c6072d60d28b41f1321ab993799d70693a3bb70bab7e58b49acc30",
			"c0de7f3035a7960450ec9a64e7835b958b0fec1ddb90cbeb0779073c0a9a8f53",
		},
	}

	verify := merkle.NewMerkleVerifier()
	var leaf_hash common.Uint256
	bys, _ := common.HexToBytes(proof.TransactionsRoot)
	leaf_hash.Deserialize(bytes.NewReader(bys))

	var root_hash common.Uint256
	bys, _ = common.HexToBytes(proof.CurBlockRoot)
	root_hash.Deserialize(bytes.NewReader(bys))

	var hashes []common.Uint256
	for _, v := range proof.TargetHashes {
		var hash common.Uint256
		bys, _ = common.HexToBytes(v)
		hash.Deserialize(bytes.NewReader(bys))
		hashes = append(hashes, hash)
	}
	res := verify.VerifyLeafHashInclusion(leaf_hash, proof.BlockHeight, hashes, root_hash, proof.CurBlockHeight+1)
	assert.Nil(t, res)

}

func TestCodeHash(t *testing.T) {
	code, _ := common.HexToBytes("")
	vmcode := vmtypes.VmCode{vmtypes.NEOVM, code}
	codehash := vmcode.AddressFromVmCode()
	assert.Contains(t, codehash.ToHexString(), "")
}

func TestTxDeserialize(t *testing.T) {
	bys, _ := common.HexToBytes("")
	var txn types.Transaction
	err := txn.Deserialize(bytes.NewReader(bys))
	assert.Nil(t, err)

	assert.Contains(t, txn.TxType, 0)
}
func TestAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	pk, err := keypair.DeserializePublicKey(pubkey)
	assert.Nil(t, err)

	ui60 := types.AddressFromPubKey(pk)
	addr := common.ToHexString(ui60[:])
	assert.Contains(t, addr, 0)
}
func TestMultiPubKeysAddress(t *testing.T) {
	pubkey, _ := common.HexToBytes("120203a4e50edc1e59979442b83f327030a56bffd08c2de3e0a404cefb4ed2cc04ca3e")
	pk, err := keypair.DeserializePublicKey(pubkey)
	assert.Nil(t, err)

	pubkey2, _ := common.HexToBytes("12020225c98cc5f82506fb9d01bad15a7be3da29c97a279bb6b55da1a3177483ab149b")
	pk2, err := keypair.DeserializePublicKey(pubkey2)
	assert.Nil(t, err)

	ui60, _ := types.AddressFromMultiPubKeys([]keypair.PublicKey{pk, pk2}, 1)
	addr := common.ToHexString(ui60[:])
	assert.Contains(t, addr, 0)
}

func TestInvokefunction(t *testing.T) {
	var funcName = "Get"
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := BuildSmartContractParamInter(builder, []interface{}{funcName, []interface{}{[]byte("key")}})
	assert.Nil(t, err)

	codeParams := builder.ToArray()
	op_verify, _ := common.HexToBytes("69")
	codeaddress, _ := common.HexToBytes("809690ff6a5244cca5e64face79914d59daef527")

	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.NEOVM,
		Code:   bytes.Join([][]byte{codeParams}, bytes.Join([][]byte{op_verify}, codeaddress)),
	})
	tx.Nonce = uint32(time.Now().Unix())

	acct := account.Open(account.WALLET_FILENAME, []byte("passwordtest"))
	acc := acct.GetDefaultAccount()
	//pubkey := keypair.SerializePublicKey(acc.PubKey())
	assert.Nil(t, err)
	hash := tx.Hash()
	sign, _ := signature.Sign(acc, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []keypair.PublicKey{acc.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})

	txbf := new(bytes.Buffer)
	err = tx.Serialize(txbf)
	assert.Nil(t, err)

	var req = map[string]interface{}{
		"Action":  "sendrawtransaction",
		"Version": "1.0.0",
		"Data":    common.ToHexString(txbf.Bytes()),
	}
	resp, err := Request("POST", req, addr+"/api/v1/transaction?preExec=1")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func BuildSmartContractParamInter(builder *neovm.ParamsBuilder, smartContractParams []interface{}) error {
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
