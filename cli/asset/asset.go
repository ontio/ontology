package asset

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ontology/account"
	. "github.com/Ontology/cli/common"
	. "github.com/Ontology/common"
	. "github.com/Ontology/core/asset"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/net/httpjsonrpc"
	"math/rand"
	"os"
	"strconv"

	"github.com/Ontology/core/transaction/utxo"
	"github.com/urfave/cli"
)

const (
	RANDBYTELEN    = 4
	REFERTXHASHLEN = 64
)

func newContractContextWithoutProgramHashes(data signature.SignableData) *contract.ContractContext {
	return &contract.ContractContext{
		Data:       data,
		Codes:      make([][]byte, 1),
		Parameters: make([][][]byte, 1),
	}
}

func openWallet(name string, passwd []byte) account.Client {
	if name == account.WalletFileName {
		fmt.Println("Using default wallet: ", account.WalletFileName)
	}
	wallet := account.Open(name, passwd)
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}
	return wallet
}

func getUintHash(programHashStr, assetHashStr string) (Uint160, Uint256, error) {
	programHashHex, err := hex.DecodeString(programHashStr)
	if err != nil {
		fmt.Println("Decoding program hash string failed")
		return Uint160{}, Uint256{}, err
	}
	var programHash Uint160
	if err := programHash.Deserialize(bytes.NewReader(programHashHex)); err != nil {
		fmt.Println("Deserialization program hash failed")
		return Uint160{}, Uint256{}, err
	}
	assetHashHex, err := hex.DecodeString(assetHashStr)
	if err != nil {
		fmt.Println("Decoding asset hash string failed")
		return Uint160{}, Uint256{}, err
	}
	var assetHash Uint256
	if err := assetHash.Deserialize(bytes.NewReader(assetHashHex)); err != nil {
		fmt.Println("Deserialization asset hash failed")
		return Uint160{}, Uint256{}, err
	}
	return programHash, assetHash, nil
}

func signTransaction(signer *account.Account, tx *transaction.Transaction) error {
	signature, err := signature.SignBySigner(tx, signer)
	if err != nil {
		fmt.Println("SignBySigner failed")
		return err
	}
	transactionContract, err := contract.CreateSignatureContract(signer.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureContract failed")
		return err
	}
	transactionContractContext := newContractContextWithoutProgramHashes(tx)
	if err := transactionContractContext.AddContract(transactionContract, signer.PubKey(), signature); err != nil {
		fmt.Println("AddContract failed")
		return err
	}
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return nil
}

func makeRegTransaction(admin, issuer *account.Account, name string, description string, value Fixed64, netWorkFee Fixed64) (string, error) {
	asset := &Asset{name, description, byte(MaxPrecision), AssetType(Share), UTXO}
	transactionContract, err := contract.CreateSignatureContract(admin.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureContract failed")
		return "", err
	}
	tx, _ := transaction.NewRegisterAssetTransaction(asset, value, issuer.PubKey(), transactionContract.ProgramHash)
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)
	tx, err = checkAndAddFees(admin.ProgramHash, tx, netWorkFee)
	if err != nil {
		return "", nil
	}
	if err := signTransaction(issuer, tx); err != nil {
		fmt.Println("sign regist transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialize registtransaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func makeIssueTransaction(issuer *account.Account, programHashStr, assetHashStr string, value Fixed64, netWorkFee Fixed64) (string, error) {
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	assetReverseHash, _ := Uint256ParseFromBytes(assetHash.ToArrayReverse())
	issueTxOutput := &utxo.TxOutput{
		AssetID:     assetReverseHash,
		Value:       value,
		ProgramHash: programHash,
	}
	outputs := []*utxo.TxOutput{issueTxOutput}
	tx, _ := transaction.NewIssueAssetTransaction(outputs)
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)
	tx, err = checkAndAddFees(issuer.ProgramHash, tx, netWorkFee)
	if err != nil {
		return "", nil
	}
	if err := signTransaction(issuer, tx); err != nil {
		fmt.Println("sign issue transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialization of issue transaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func makeTransferTransaction(signer *account.Account, programHashStr, assetHashStr string, value Fixed64, netWorkFee Fixed64) (string, error) {
	inputs := []*utxo.UTXOTxInput{}
	outputs := []*utxo.TxOutput{}
	// get user id & asset id
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	reverseHash, _ := Uint256ParseFromBytes(assetHash.ToArrayReverse())
	var tx *transaction.Transaction
	inputs, outputs, err = calcUtxoByRpc(signer.ProgramHash, programHash, reverseHash, value, 0, false)
	if err != nil {
		return "", err
	}
	tx, _ = transaction.NewTransferAssetTransaction(inputs, outputs)
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)
	if err := signTransaction(signer, tx); err != nil {
		fmt.Println("sign transfer transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialization of transfer transaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func checkAndAddFees(Spender Uint160, Tx *transaction.Transaction, networkFee Fixed64) (*transaction.Transaction, error) {
	return Tx, nil
}

func calcUtxoByRpc(spender Uint160, toAddr Uint160, assetID Uint256, value Fixed64, fee Fixed64, throw bool) ([]*utxo.UTXOTxInput, []*utxo.TxOutput, error) {
	type UTXOUnspentInfo struct {
		Txid  string
		Index uint32
		Value int64
	}
	//get spender utxos
	b_buf := new(bytes.Buffer)
	assetID.Serialize(b_buf)
	assetIDHex := hex.EncodeToString(b_buf.Bytes())
	resp, err := httpjsonrpc.Call(Address(), "getunspendoutput", 0, []interface{}{ToHexString(spender.ToArray()), assetIDHex})
	if err != nil {
		fmt.Println("HTTP JSON call failed")
		return nil, nil, err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return nil, nil, err
	}
	var unspend []interface{}
	switch res := r["result"].(type) {
	case []interface{}:
		unspend = res
	default:
		return nil, nil, errors.New(fmt.Sprintf("[calcUtxoByRpc] failed with invalid value returned with value=%s\n",res))
	}
	//calc inputs and outputs
	inputs := []*utxo.UTXOTxInput{}
	outputs := []*utxo.TxOutput{}
	if value != 0 && throw == false {
		transferTxOutput := &utxo.TxOutput{
			AssetID:     assetID,
			Value:       value,
			ProgramHash: toAddr,
		}
		outputs = append(outputs, transferTxOutput)
	}

	expected := value + fee
	for _, v := range unspend {
		var unspentUtxo UTXOUnspentInfo
		temp := v.(map[string]interface{})
		if unspentUtxo.Value, err = strconv.ParseInt(temp["Value"].(string), 10, 64); err != nil {
			return nil, nil, err
		}
		if index_, err := strconv.ParseInt(temp["Index"].(string), 10, 64); err != nil {
			return nil, nil, err
		} else {
			unspentUtxo.Index = uint32(index_)
		}
		unspentUtxo.Txid = temp["Txid"].(string)
		h := unspentUtxo.Txid
		referIndex := unspentUtxo.Index
		b, _ := hex.DecodeString(h)
		var referHash Uint256
		referHash.Deserialize(bytes.NewReader(b))
		value := Fixed64(unspentUtxo.Value)
		if value == expected {
			transferUTXOInput := &utxo.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			expected = 0
			inputs = append(inputs, transferUTXOInput)
			break
		} else if value > expected {
			transferUTXOInput := &utxo.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			inputs = append(inputs, transferUTXOInput)
			getChangeOutput := &utxo.TxOutput{
				AssetID:     assetID,
				Value:       value - expected,
				ProgramHash: spender,
			}
			expected = 0
			outputs = append(outputs, getChangeOutput)
			break
		} else if value < expected {
			transferUTXOInput := &utxo.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			expected -= value
			inputs = append(inputs, transferUTXOInput)
			if expected == 0 {
				break
			}
		}
	}
	if expected != 0 {
		return nil, nil, errors.New(fmt.Sprintf("transfer failed, ammount is not enough, expected is %d\n", expected))
	}
	return inputs, outputs, nil
}

func assetAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	var funcName string
	if c.Bool("reg") == true {
		funcName = "reg"
	}
	if c.Bool("issue") == true {
		funcName = "issue"
	}
	if c.Bool("transfer") == true {
		funcName = "transfer"
	}
	if !c.Bool("reg") && !c.Bool("issue") && !c.Bool("transfer") {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	wallet := openWallet(c.String("wallet"), WalletPassword(c.String("password")))
	admin, _ := wallet.GetDefaultAccount()
	value := c.Int64("value")
	netWorkFee := c.Int64("netWorkFee")
	var txHex string
	var err error
	switch funcName {
	case "reg":
		name := c.String("name")
		if name == "" {
			rbuf := make([]byte, RANDBYTELEN)
			rand.Read(rbuf)
			name = "Ontology-" + ToHexString(rbuf)
		}
		issuer := admin
		description := "description"
		txHex, err = makeRegTransaction(admin, issuer, name, description, Fixed64(value), Fixed64(netWorkFee))
	case "issue":
		asset := c.String("asset")
		to := c.String("to")
		if asset == "" || to == "" {
			fmt.Println("missing flag [--asset] or [--to]")
			return nil
		}
		txHex, err = makeIssueTransaction(admin, to, asset, Fixed64(value), Fixed64(netWorkFee))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	case "transfer":
		asset := c.String("asset")
		to := c.String("to")
		if asset == "" || to == "" {
			fmt.Println("missing flag [--asset] or [--to]")
			return nil
		}
		txHex, err = makeTransferTransaction(admin, to, asset, Fixed64(value), Fixed64(netWorkFee))
		if err != nil {
			fmt.Println(err)
			return nil
		}
	default:
		cli.ShowSubcommandHelp(c)
		return nil
	}
	resp, err := httpjsonrpc.Call(Address(), "sendrawtransaction", 0, []interface{}{txHex})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	FormatOutput(resp)

	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "asset",
		Usage:       "asset registration, issuance and transfer",
		Description: "With nodectl asset, you could control assert through transaction.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "reg, r",
				Usage: "regist a new kind of asset",
			},
			cli.BoolFlag{
				Name:  "issue, i",
				Usage: "issue asset that has been registered",
			},
			cli.BoolFlag{
				Name:  "transfer, t",
				Usage: "transfer asset",
			},
			cli.StringFlag{
				Name:  "wallet, w",
				Usage: "wallet name",
				Value: account.WalletFileName,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
			},
			cli.StringFlag{
				Name:  "asset, a",
				Usage: "uniq id for asset",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "asset name",
			},
			cli.StringFlag{
				Name:  "to",
				Usage: "asset to whom",
			},
			cli.StringFlag{
				Name:  "referTxID",
				Usage: "referTxID of utxo",
			},
			cli.StringFlag{
				Name:  "index",
				Usage: "index of utxo",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "asset ammount",
			},
			cli.Int64Flag{
				Name:  "netWorkFee, f",
				Usage: "netWorkFee ammount",
			},
		},
		Action: assetAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "asset")
			return cli.NewExitError("", 1)
		},
	}
}