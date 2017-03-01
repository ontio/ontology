package main

import (
	"GoOnchain/common/log"
	"GoOnchain/core/ledger"
	"GoOnchain/core/store"
	"GoOnchain/core/transaction"
	"GoOnchain/crypto"
	"GoOnchain/net"
	"GoOnchain/net/httpjsonrpc"
	"fmt"
	"runtime"
	"time"
	//"GoOnchain/consensus/dbft"
	. "GoOnchain/client"
	. "GoOnchain/common"
	"GoOnchain/consensus/dbft"
	. "GoOnchain/core/asset"
	"GoOnchain/core/contract"
	"GoOnchain/core/signature"
	"crypto/sha256"
	"GoOnchain/core/validation"
	"os"
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
	fmt.Printf("Node version: %s\n", Version)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 0. Client Set                                                      ***")
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
	if localclient == nil{
		fmt.Println("Can't get local client.")
		os.Exit(1)
	}
	issuer,err:= localclient.GetDefaultAccount()
	if err != nil {
		fmt.Println(err)
	}
	admin := issuer

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Set Miner                                                     ***")
	fmt.Println("//**************************************************************************")
	miner := []*crypto.PubKey{}
	miner = append(miner, getMiner().PublicKey)
	ledger.StandbyMiners = miner
	fmt.Println("miner1.PublicKey", issuer.PublicKey)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 3. BlockChain init                                                 ***")
	fmt.Println("//**************************************************************************")
	sampleBlockchain := InitBlockChain()
	ledger.DefaultLedger.Blockchain = &sampleBlockchain

	time.Sleep(2 * time.Second)
	neter := net.StartProtocol()
	time.Sleep(2 * time.Second)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 5. Start DBFT Services                                             ***")
	fmt.Println("//**************************************************************************")
	dbftServices := dbft.NewDbftService(localclient, "/opt/dbft", neter)
	dbftServices.Start()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 6. Create Transaction by client.Account                            ***")
	fmt.Println("//**************************************************************************")



	fmt.Println("//**************************************************************************")
	fmt.Println("//*** Init Complete                                                      ***")
	fmt.Println("//**************************************************************************")
	go httpjsonrpc.StartServer()

	time.Sleep(2 * time.Second)
	httpjsonrpc.StartClient()
	// Modules start sample
	//ledger.Start(net.NetToLedgerCh <-chan *Msg, net.LedgerToNetCh chan<- *Msg)
	//consensus.Start(net.NetToConsensusCh <-chan *Msg, net.ConsensusToNetCh chan<- *Msg)

	if os.Getenv("CLIENT_NAME") == "c4" {
		tx := sampleTransaction(issuer, admin)
		neter.Xmit(tx)
	}

	for {
		time.Sleep(2 * time.Second)
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
	fmt.Println("//*** 3. Generate [Asset] Test                                           ***")
	fmt.Println("//**************************************************************************")
	a1 := SampleAsset()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 4. [controllerPGM] Generate Test                                   ***")
	fmt.Println("//**************************************************************************")
	controllerPGM, _ := contract.CreateSignatureContract(admin.PubKey())

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 6. Generate [Transaction] Test                                     ***")
	fmt.Println("//**************************************************************************")
	ammount := Fixed64(10)
	tx, _ := transaction.NewAssetRegistrationTransaction(a1, &ammount, issuer.PubKey(), &controllerPGM.ProgramHash)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 7. Generate [signature],[sign],set transaction [Program]           ***")
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
	fmt.Println("//*** 8. Transaction [Validation]                                       ***")
	fmt.Println("//**************************************************************************")
	//1.validate transaction content
	err = validation.VerifyTransaction(tx, ledger.DefaultLedger, nil)
	if err != nil {
		fmt.Println("Transaction Verify error.", err)
	} else {
		fmt.Println("Transaction Verify Normal Completed.")
	}
	//2.validate transaction signdate
	_,err = validation.VerifySignature(tx, issuer.PubKey(), signdate)
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

func OpenClientAndGetAccount() *Client {
	//CreateClient( "wallet.db3", []byte("\x12\x34\x56") )
	//cl := OpenClient( "wallet.db3", []byte("\x12\x34\x56") )
	clientName := os.Getenv("CLIENT_NAME")
	if  clientName== ""{
		fmt.Printf("Please Check your client's ENV SET, which schould be c1,c2,c3,c4. Now is %s\n",clientName)
		return nil
	}
	//c1 := CreateClient("wallet1.db3", []byte("\x12\x34\x56"))
	//c2 := CreateClient("wallet2.db3", []byte("\x12\x34\x56"))
	//c3 := CreateClient("wallet3.db3", []byte("\x12\x34\x56"))
	//c4 := CreateClient("wallet4.db3", []byte("\x12\x34\x56"))
	c1 := OpenClient("wallet1.db3", []byte("\x12\x34\x56"))
	c2 := OpenClient("wallet2.db3", []byte("\x12\x34\x56"))
	c3 := OpenClient("wallet3.db3", []byte("\x12\x34\x56"))
	c4 := OpenClient("wallet4.db3", []byte("\x12\x34\x56"))

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

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func getMiner()*Account{
	c4 := OpenClient("wallet4.db3", []byte("\x12\x34\x56"))
	account,err:= c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}