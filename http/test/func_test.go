package test

import (
	"testing"
	"bytes"
	"fmt"
	"os"
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	"github.com/Ontology/crypto"
	"math/big"
	"github.com/Ontology/smartcontract/service/native/states"
	"github.com/Ontology/common/serialization"
)

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
	crypto.SetAlg("")
	pubkey, _ := common.HexToBytes("0399b851bc2cd05506d6821d4bc5a92139b00ac4bc7399cd9ca0aac86a468d1c05")
	pk, err := crypto.DecodePoint(pubkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	ui60 := types.AddressFromPubKey(pk)
	addr := common.ToHexString(ui60.ToArray())
	fmt.Println(addr)
	fmt.Println(ui60.ToBase58())
}
func TestTransfer(t *testing.T)  {
	contract := "ff00000000000000000000000000000000000001"

	ct, _ := common.HexToBytes(contract)
	ctu, _ := common.Uint160ParseFromBytes(ct)
	from := "0121dca8ffcba308e697ee9e734ce686f4181658"

	f, _ := common.HexToBytes(from)
	fu, _ := common.Uint160ParseFromBytes(f)
	to := "01c6d97beeb85c7fef8cea8edd564f52c0236fb1"

	tt, _ := common.HexToBytes(to)
	tu, _ := common.Uint160ParseFromBytes(tt)

	var sts []*states.State
	sts = append(sts, &states.State{
		From: fu,
		To: tu,
		Value: big.NewInt(100),
	})
	transfers := new(states.Transfers)
	fmt.Println("ctu:", ctu)
	transfers.Params = append(transfers.Params, &states.TokenTransfer{
		Contract: ctu,
		States: sts,
	})

	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, []byte("Token.Common.Transfer")); err != nil {
		fmt.Println("Serialize transfer falg error.")
		os.Exit(1)
	}
	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	fmt.Println(common.ToHexString(bf.Bytes()))

}