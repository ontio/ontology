package test

import (
	"DNA/client"
	. "DNA/common"
	. "DNA/core/asset"
	"DNA/core/contract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"DNA/net/httpjsonrpc"
	"DNA/utility"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

const (
	RANDBYTELEN = 32
	ASSETPREFIX = "DNA :"
)

var usage = `run sample routines`

var flags = []string{"tx", "num", "nosign"}

func main(args []string, p utility.Param) (err error) {
	if p.Tx {
		issuer, err := client.NewAccount()
		if err != nil {
			return err
		}
		admin := issuer

		rbuf := make([]byte, RANDBYTELEN)
		rand.Read(rbuf)
		assetid := Uint256(sha256.Sum256(rbuf))

		var index int64
		for index = 0; index < p.TxNum; index++ {
			tx := sampleTransaction(issuer, admin, assetid, index, p.NoSign)
			buf := new(bytes.Buffer)
			err = tx.Serialize(buf)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
			resp, err := httpjsonrpc.Call(utility.Address(p.Ip, p.Port), "sendsampletransaction", p.RPCID, []interface{}{buf.Bytes()})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
			utility.FormatOutput(resp)
		}
	}

	return nil
}

func sampleTransaction(issuer, admin *client.Account, assetid Uint256, index int64, nosign bool) *transaction.Transaction {
	// generate asset
	a := SampleAsset(index)
	// generate controllerPGM
	controllerPGM, _ := contract.CreateSignatureContract(admin.PubKey())
	// generate transaction
	ammount := Fixed64(10)
	tx, _ := transaction.NewRegisterAssetTransaction(a, ammount, issuer.PubKey(), controllerPGM.ProgramHash)
	if nosign {
		return tx
	}
	// generate signature
	signdate, err := signature.SignBySigner(tx, issuer)
	if err != nil {
		fmt.Println(err, "signdate SignBySigner failed")
	}
	// create & add contract
	transactionContract, _ := contract.CreateSignatureContract(issuer.PubKey())
	transactionContractContext := contract.NewContractContext(tx)
	transactionContractContext.AddContract(transactionContract, issuer.PublicKey, signdate)
	// get ContractContext Programs & set into transaction
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return tx
}

func SampleAsset(index int64) *Asset {
	name := ASSETPREFIX + strconv.FormatInt(index, 10)
	asset := Asset{name, byte(0x00), AssetType(Share), UTXO}
	return &asset
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: main}
