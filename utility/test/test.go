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
	_"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"DNA/common/log"
	"time"
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
		//assetid := Uint256(sha256.Sum256(rbuf))
		var regHash Uint256
		var issueHash Uint256
		var index int64
		for index = 0; index < p.TxNum; index++ {
			// RegisterAsset
			{
				tx := sampleTransaction(issuer, admin, index, rbuf, p.NoSign)
				buf := new(bytes.Buffer)
				err = tx.Serialize(buf)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return err
				}
				regHash = tx.Hash()
				resp, err := httpjsonrpc.Call(utility.Address(p.Ip, p.Port), "sendsampletransaction", p.RPCID, []interface{}{buf.Bytes()})
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return err
				}

				utility.FormatOutput(resp)
			}
			time.Sleep(5 * time.Second)
			//IssueAsset
			{
				tx := sampleTransactionIssue(admin, regHash)
				buf := new(bytes.Buffer)
				err = tx.Serialize(buf)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return err
				}
				issueHash = tx.Hash()
				resp, err := httpjsonrpc.Call(utility.Address(p.Ip, p.Port), "sendsampletransaction", p.RPCID, []interface{}{buf.Bytes()})
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return err
				}
				utility.FormatOutput(resp)

			}
			time.Sleep(5 * time.Second)
			//TransferAsset
			{
				tx := sampleTransactionTransfer(issuer, admin, regHash, issueHash)
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
		var out bytes.Buffer
		out.Write([]byte(fmt.Sprintf("hash issue:%x\n",regHash)))
		out.Write([]byte(fmt.Sprintf("hash transfer:%x\n",issueHash)))
		out.Write([]byte("\n"))
		_, err = out.WriteTo(os.Stdout)
	}

	return nil
}

func sampleTransaction(issuer, admin *client.Account, index int64,buf []byte, nosign bool) *transaction.Transaction {
	// generate asset
	a := SampleAsset(index,buf)
	// generate controllerPGM
	controllerPGM, _ := contract.CreateSignatureContract(admin.PubKey())
	// generate transaction
	ammount := Fixed64(1000)
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

func sampleTransactionIssue(admin *client.Account,hash Uint256) *transaction.Transaction {
	// generate transaction
	tx, _ := transaction.NewIssueAssetTransaction()
	// generate signature
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		fmt.Println(err, "signdate SignBySigner failed")
	}

	// create & add contract
	transactionContract, _ := contract.CreateSignatureContract(admin.PublicKey)
	transactionContractContext := contract.NewContractContext(tx)
	transactionContractContext.AddContract(transactionContract, admin.PublicKey, signdate)

	//UTXO　Output generate.
	temp,err := admin.PublicKey.EncodePoint(true)
	if err !=nil{
		log.Debug("EncodePoint error.")
	}
	hashx ,err  := ToCodeHash(temp)
	if err != nil {
		log.Debug("TocodeHash hash error.")
	}
	issueTxOutput := &transaction.TxOutput{
		AssetID:hash,
		Value: Fixed64(100),
		ProgramHash:hashx,
	}
	tx.Outputs = append(tx.Outputs,issueTxOutput)

	// get ContractContext Programs & set into transaction
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return tx
}

func sampleTransactionTransfer(toUser, admin *client.Account,regHash Uint256,issueHash Uint256) *transaction.Transaction {
	// generate transaction
	tx, _ := transaction.NewTransferAssetTransaction()
	// generate signature
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		fmt.Println(err, "signdate SignBySigner failed")
	}
	// create & add contract
	transactionContract, _ := contract.CreateSignatureContract(admin.PublicKey)
	transactionContractContext := contract.NewContractContext(tx)
	transactionContractContext.AddContract(transactionContract, admin.PublicKey, signdate)

	//UTXO  INPUT Generate.
	transferUTXOInput := &transaction.UTXOTxInput{
		ReferTxID:issueHash,
		//The index of output in the referTx output list
		ReferTxOutputIndex:uint16(0),
	}
	tx.UTXOInputs = append(tx.UTXOInputs,transferUTXOInput)

	//UTXO　Output generate.
	temp,err := toUser.PublicKey.EncodePoint(true)
	if err !=nil{
		log.Debug("EncodePoint error.")
	}
	hashx ,err  := ToCodeHash(temp)
	if err != nil {
		log.Debug("TocodeHash hash error.")
	}
	issueTxOutput := &transaction.TxOutput{
		AssetID:regHash,
		Value: Fixed64(100),
		ProgramHash:hashx,
	}
	tx.Outputs = append(tx.Outputs,issueTxOutput)

	// get ContractContext Programs & set into transaction
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return tx
}


func SampleAsset( index int64,buf []byte) *Asset {
	name := ASSETPREFIX + string(buf) + strconv.FormatInt(index,10)
	asset := Asset{name, byte(0x00), AssetType(Share), UTXO}
	return &asset
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: main}
