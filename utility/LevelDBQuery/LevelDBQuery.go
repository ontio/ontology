package main

import (
	"DNA/crypto"
	"DNA/core/ledger"
	"DNA/core/store"
	."DNA/common"
	"DNA/common/log"
	"encoding/hex"
	"os"
	"fmt"
)

const (
	path string = "./Log"
)

var Usage = func() {
	fmt.Println("Usage:")
	fmt.Println("  dump");
	fmt.Println("  block <hash>");
	fmt.Println("  tx <txid>");
	fmt.Println("  asset <assetid>");
}

func DumpDB() {
	fd,_:=os.OpenFile("dump.txt",os.O_RDWR|os.O_CREATE,0644)

	i := 0
	for (true) {
		hash, err := ledger.DefaultLedger.Store.GetBlockHash(uint32(i))
		if err != nil {
			fmt.Println("Dump() ok.")
			return
		}

		block,err := ledger.DefaultLedger.Store.GetBlock(hash)
		if err != nil {
			fmt.Println("GetBlock() err!")
			return
		}

		fd.Write( []byte(fmt.Sprintf("Block height: %d\n", block.Blockdata.Height)) )
		h := block.Blockdata.Hash()
		fd.Write( []byte(fmt.Sprintf( "Block hash: %x\n", h.ToArray() )) )
		fd.Write( []byte(fmt.Sprintf( "Block timestamp: %d\n", block.Blockdata.Timestamp )) )
		fd.Write( []byte(fmt.Sprintf( "Block transactionsRoot :%x\n", block.Blockdata.TransactionsRoot )) )
		fd.Write( []byte(fmt.Sprintf( "Tx Len: %d\n", len(block.Transcations) )) )

		for k:=0; k<len(block.Transcations); k++ {
			txhash := block.Transcations[k].Hash()
			fd.Write( []byte(fmt.Sprintf( "Tx hash: %x\n", txhash.ToArray() )) )
			fd.Write( []byte(fmt.Sprintf( "Tx Type: %x\n", block.Transcations[k].TxType )) )
		}

		fd.Write( []byte("\n") )
		i ++
	}

	fd.Close()
}


func GetBlock(hash string) {
	bhash,err := hex.DecodeString(hash)
	uhash,err := Uint256ParseFromBytes( bhash )
	if err != nil {
		fmt.Printf( "bhash len: %d\n", len(bhash) )
		fmt.Printf( "Uint256ParseFromBytes err: %s\n", err )
	}

	block,err := ledger.DefaultLedger.Store.GetBlock(uhash)
	if err == nil {
		fmt.Printf( "hash: %s\n", hash )
		fmt.Printf( "height: %d\n", block.Blockdata.Height )
		fmt.Printf( "Tx Len: %d\n", len(block.Transcations) )

		for k:=0; k<len(block.Transcations); k++ {
			txhash := block.Transcations[k].Hash()
			fmt.Printf( "Tx hash: %x\n", txhash.ToArray() )
			fmt.Printf( "Tx Type: %x\n", block.Transcations[k].TxType )
			fmt.Printf( "\n" )
		}
	} else {
		fmt.Printf( "err: %s\n", err )
	}
}

func GetTx(txid string) {
	bhash,err := hex.DecodeString(txid)
	uhash,err := Uint256ParseFromBytes( bhash )
	if err != nil {
		fmt.Printf( "bhash len: %d\n", len(bhash) )
		fmt.Printf( "Uint256ParseFromBytes err: %s\n", err )
	}

	tx,err := ledger.DefaultLedger.Store.GetTransaction(uhash)
	if err == nil {
		fmt.Printf( "txid: %s\n", txid )
		fmt.Printf( "tx.TxType: %x\n", tx.TxType )
	} else {
		fmt.Printf( "err: %s\n", err )
	}
}


func GetAsset(assetid string) {
	bhash,err := hex.DecodeString(assetid)
	uhash,err := Uint256ParseFromBytes( bhash )
	if err != nil {
		fmt.Printf( "bhash len: %d\n", len(bhash) )
		fmt.Printf( "Uint256ParseFromBytes err: %s\n", err )
	}

	asset,err := ledger.DefaultLedger.Store.GetAsset(uhash)
	if err == nil {
		fmt.Printf( "txid: %s\n", assetid )
		asset.ID
		fmt.Printf( "asset.TxType: %x\n", asset.AssetType )
		fmt.Printf( "asset.Name: %s\n", asset.Name )
		fmt.Printf( "asset.RecordType: %x\n", asset.RecordType )
		fmt.Printf( "asset.Precision: %x\n", asset.Precision )
	} else {
		fmt.Printf( "err: %s\n", err )
	}
}

func main() {
	crypto.SetAlg(crypto.P256R1)
	log.CreatePrintLog(path)

	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = store.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)

	args := os.Args
	if args == nil || len(args)<2{
		Usage()
		return
	}

	cmd := args[1]
	if cmd == "dump" {
		DumpDB()
	} else if cmd == "block" {
		blockhash := args[2]
		GetBlock(blockhash)
	} else if cmd == "tx" {
		txid := args[2]
		GetTx(txid)
	}  else if cmd == "asset" {
		assetid := args[2]
		GetAsset(assetid)
	}

}

