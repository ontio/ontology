package main

import (
	. "GoOnchain/client"
	. "GoOnchain/common"
	"GoOnchain/common/log"
	"GoOnchain/config"
	"GoOnchain/consensus/dbft"
	. "GoOnchain/core/asset"
	"GoOnchain/core/contract"
	"GoOnchain/core/ledger"
	"GoOnchain/core/signature"
	"GoOnchain/core/store"
	"GoOnchain/core/transaction"
	"GoOnchain/core/validation"
	"GoOnchain/crypto"
	"GoOnchain/net"
	"GoOnchain/net/httpjsonrpc"
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	// The number of the CPU cores for parallel optimization,TODO set from config file
	NCPU = 4
)

var Version string

func init() {
	runtime.GOMAXPROCS(NCPU)
	var path string = "./Log/"
	log.CreatePrintLog(path)
}

func main() {
	var path string = "./Log/"
	log.CreatePrintLog(path)
	fmt.Printf("Node version: %s\n", Version)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 0. Client open                                                     ***")
	fmt.Println("//**************************************************************************")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = store.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(crypto.P256R1)
	fmt.Println("  Client set completed. Test Start...")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 1. Generate [Account]                                              ***")
	fmt.Println("//**************************************************************************")
	localclient := OpenClientAndGetAccount()
	if localclient == nil {
		fmt.Println("Can't get local client.")
		os.Exit(1)
	}

	issuer, err := localclient.GetDefaultAccount()
	if err != nil {
		fmt.Println(err)
	}
	//admin := issuer

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Set Miner                                                     ***")
	fmt.Println("//**************************************************************************")
	miner := []*crypto.PubKey{}
	miner = append(miner, getMiner1().PublicKey)
	miner = append(miner, getMiner4().PublicKey)
	miner = append(miner, getMiner3().PublicKey)
	miner = append(miner, getMiner2().PublicKey)
	ledger.StandbyMiners = miner
	fmt.Println("miner1.PublicKey", issuer.PublicKey)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 3. BlockChain init                                                 ***")
	fmt.Println("//**************************************************************************")
	sampleBlockchain := InitBlockChain()
	ledger.DefaultLedger.Blockchain = &sampleBlockchain

	time.Sleep(2 * time.Second)
	neter, noder := net.StartProtocol()
	httpjsonrpc.RegistRpcNode(noder)
	time.Sleep(1 * time.Minute)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 5. Start DBFT Services                                             ***")
	fmt.Println("//**************************************************************************")
	dbftServices := dbft.NewDbftService(localclient, "logdbft", neter)
	httpjsonrpc.RegistDbftService(dbftServices)
	go dbftServices.Start()
	time.Sleep(5 * time.Second)
	fmt.Println("DBFT Services start completed.")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** Init Complete                                                      ***")
	fmt.Println("//**************************************************************************")
	//go httpjsonrpc.StartRPCServer()
	//go httpjsonrpc.StartLocalServer()

	time.Sleep(2 * time.Second)
	// if config.Parameters.MinerName == "c4" {
	// 	time.Sleep(2 * time.Second)
	// 	tx := sampleTransaction(issuer, admin)
	// 	fmt.Println("//**************************************************************************")
	// 	fmt.Println("//*** transaction gen complete, neter Xmit start                         ***")
	// 	fmt.Println("//**************************************************************************")
	// 	neter.Xmit(tx)
	// 	time.Sleep(10 * time.Second)
	// 	fmt.Println("//**************************************************************************")
	// 	fmt.Println("//*** neter Xmit completed                                               ***")
	// 	fmt.Println("//**************************************************************************")
	// 	//for {
	// 		fmt.Println("ledger.DefaultLedger.Blockchain.BlockHeight", ledger.DefaultLedger.Blockchain.BlockHeight)
	// 		genesisBlockHash, _ := ledger.DefaultLedger.Store.GetBlockHash(0)
	// 		fmt.Println("gensisBlockGet =", genesisBlockHash)
	// 		firstblock, _ := ledger.DefaultLedger.Store.GetBlockHash(1)
	// 		fmt.Println("FirstBlockGet =", firstblock)
	// 		time.Sleep(10 * time.Second)
	// 	//}

	// }

	for {
		log.Debug("ledger.DefaultLedger.Blockchain.BlockHeight= ", ledger.DefaultLedger.Blockchain.BlockHeight)
		time.Sleep(15 * time.Second)
	}
}
func InitBlockChain() ledger.Blockchain {
	blockchain, err := ledger.NewBlockchainWithGenesisBlock()
	if err != nil {
		fmt.Println(err, "  BlockChain generate failed")
	}
	fmt.Println("  BlockChain generate completed. Func test Start...")
	return *blockchain
}

func sampleTransaction(issuer *Account, admin *Account) *transaction.Transaction {
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-1. Generate [Asset] Test                                           ***")
	fmt.Println("//**************************************************************************")
	a1 := SampleAsset()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-2. [controllerPGM] Generate Test                                   ***")
	fmt.Println("//**************************************************************************")
	controllerPGM, _ := contract.CreateSignatureContract(admin.PubKey())

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-3. Generate [Transaction] Test                                     ***")
	fmt.Println("//**************************************************************************")
	ammount := Fixed64(10)
	tx, _ := transaction.NewAssetRegistrationTransaction(a1, &ammount, issuer.PubKey(), &controllerPGM.ProgramHash)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-4. Generate [signature],[sign],set transaction [Program]           ***")
	fmt.Println("//**************************************************************************")

	//1.Transaction [Contract]
	transactionContract, _ := contract.CreateSignatureContract(issuer.PubKey())
	//2.Transaction Signdate
	signdate, err := signature.SignBySigner(tx, issuer)
	if err != nil {
		fmt.Println(err, "signdate SignBySigner failed")
	}
	//3.Transaction [contractContext]
	fmt.Println("11111 transactionContract.Code", transactionContract.Code)
	fmt.Println("11111 transactionContract.Parameters", transactionContract.Parameters)
	fmt.Println("11111 transactionContract.ProgramHash", transactionContract.ProgramHash)
	transactionContractContext := contract.NewContractContext(tx)
	//4.add  Contract , public key, signdate to ContractContext
	transactionContractContext.AddContract(transactionContract, issuer.PublicKey, signdate)
	fmt.Println("22222 transactionContract.Code=", transactionContractContext.Codes)
	fmt.Println("22222 ", transactionContractContext.GetPrograms()[0])

	//5.get ContractContext Programs & setinto transaction
	tx.SetPrograms(transactionContractContext.GetPrograms())

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-5. Transaction [Validation]                                       ***")
	fmt.Println("//**************************************************************************")
	//1.validate transaction content
	err = validation.VerifyTransaction(tx, ledger.DefaultLedger, nil)
	if err != nil {
		fmt.Println("Transaction Verify error.", err)
	} else {
		fmt.Println("Transaction Verify Normal Completed.")
	}
	//2.validate transaction signdate
	_, err = validation.VerifySignature(tx, issuer.PubKey(), signdate)
	if err != nil {
		fmt.Println("Transaction Signature Verify error.", err)
	} else {
		fmt.Println("Transaction Signature Verify Normal Completed.")
	}
	return tx
}
func SampleAsset() *Asset {
	var x string = "Onchain"
	a1 := Asset{Uint256(sha256.Sum256([]byte("a"))), x, byte(0x00), AssetType(Share), UTXO}
	fmt.Println("  Asset generate complete. Func test Start...")
	return &a1
}

func OpenClientAndGetAccount() Client {
	clientName := config.Parameters.MinerName
	fmt.Printf("The Miner name is %s\n", clientName)
	if clientName == "" {
		fmt.Printf("Miner name not be set at config file protocol.json, which schould be c1,c2,c3,c4. Now is %s\n", clientName)
		return nil
	}
	var c1 Client
	var c2 Client
	var c3 Client
	var c4 Client

	if fileExisted("wallet1.txt") {
		c1 = OpenClient("wallet1.txt", []byte("\x12\x34\x56"))
	} else {
		c1 = CreateClient("wallet1.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet2.txt") {
		c2 = OpenClient("wallet2.txt", []byte("\x12\x34\x56"))
	} else {
		c2 = CreateClient("wallet2.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet3.txt") {
		c3 = OpenClient("wallet3.txt", []byte("\x12\x34\x56"))
	} else {
		c3 = CreateClient("wallet3.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet4.txt") {
		c4 = OpenClient("wallet4.txt", []byte("\x12\x34\x56"))
	} else {
		c4 = CreateClient("wallet4.txt", []byte("\x12\x34\x56"))
	}

	//ac,_ := cl.GetDefaultAccount()
	//fmt.Printf("PrivateKey: %x\n", ac.PrivateKey)
	//fmt.Printf("PublicKeyHash: %x\n", ac.PublicKeyHash.ToArray())
	//fmt.Printf("PublicKeyAddress: %s\n", ac.PublicKeyHash.ToAddress())
	switch clientName {
	case "c1":
		return c1
	case "c2":
		return c2
	case "c3":
		return c3
	case "c4":
		return c4
	default:
		fmt.Printf("Please Check your client's ENV SET, which schould be c1,c2,c3,c4. Now is %s.\n", clientName)
		return nil
	}
}

func fileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func getMiner1() *Account {
	c4 := OpenClient("wallet1.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner2() *Account {
	c4 := OpenClient("wallet2.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner3() *Account {
	c4 := OpenClient("wallet3.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner4() *Account {
	c4 := OpenClient("wallet4.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
