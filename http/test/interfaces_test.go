package test

import (
	"testing"
	"bytes"
	"fmt"
	"os"
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	"github.com/Ontology/crypto"
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
}
