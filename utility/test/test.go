package test

import (
	"DNA/client"
	. "DNA/common"
	. "DNA/core/asset"
	"DNA/core/contract"
	"DNA/core/ledger"
	"DNA/core/signature"
	"DNA/core/transaction"
	"DNA/core/validation"
	"DNA/net/httpjsonrpc"
	"DNA/utility"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
)

var usage = `run sample programs`

var flags = []string{"tx"}

func main(args []string, p utility.Param) (err error) {
	if p.Tx {
		issuer, err := client.NewAccount()
		if err != nil {
			return err
		}
		admin := issuer
		tx := sampleTransaction(issuer, admin)
		buf := new(bytes.Buffer)
		err = tx.Serialize(buf)
		if err != nil {
			return err
		}
		resp, err := httpjsonrpc.Call(utility.Address(p.Ip, p.Port), "sendsampletransaction", p.RPCID, []interface{}{buf.Bytes()})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		utility.FormatOutput(resp)
	}

	return nil
}

func sampleTransaction(issuer, admin *client.Account) *transaction.Transaction {
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-1. Generate [Asset] Test                                         ***")
	fmt.Println("//**************************************************************************")
	a := SampleAsset()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-2. [controllerPGM] Generate Test                                 ***")
	fmt.Println("//**************************************************************************")
	controllerPGM, _ := contract.CreateSignatureContract(admin.PubKey())

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-3. Generate [Transaction] Test                                   ***")
	fmt.Println("//**************************************************************************")
	ammount := Fixed64(10)
	tx, _ := transaction.NewAssetRegistrationTransaction(a, &ammount, issuer.PubKey(), &controllerPGM.ProgramHash)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-4. Generate [signature],[sign],set transaction [Program]         ***")
	fmt.Println("//**************************************************************************")

	// 1.Transaction [Contract]
	transactionContract, _ := contract.CreateSignatureContract(issuer.PubKey())
	// 2.Transaction Signdate
	signdate, err := signature.SignBySigner(tx, issuer)
	if err != nil {
		fmt.Println(err, "signdate SignBySigner failed")
	}
	// 3.Transaction [contractContext]
	transactionContractContext := contract.NewContractContext(tx)
	// 4.add  Contract , public key, signdate to ContractContext
	transactionContractContext.AddContract(transactionContract, issuer.PublicKey, signdate)

	// 5.get ContractContext Programs & setinto transaction
	tx.SetPrograms(transactionContractContext.GetPrograms())

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** A-5. Transaction [Validation]                                      ***")
	fmt.Println("//**************************************************************************")
	// 1.validate transaction content
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
	a := Asset{Uint256(sha256.Sum256([]byte("a"))), x, byte(0x00), AssetType(Share), UTXO}
	fmt.Println("Asset generate complete. Func test Start...")
	return &a
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: main}
