package asset

import (
	"DNA/account"
	. "DNA/cli/common"
	. "DNA/common"
	. "DNA/core/asset"
	"DNA/core/contract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"DNA/net/httpjsonrpc"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"

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

func makeRegTransaction(admin, issuer *account.Account, name string, description string, value Fixed64) (string, error) {
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

func makeIssueTransaction(issuer *account.Account, programHashStr, assetHashStr string, value Fixed64) (string, error) {
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	issueTxOutput := &transaction.TxOutput{
		AssetID:     assetHash,
		Value:       value,
		ProgramHash: programHash,
	}
	outputs := []*transaction.TxOutput{issueTxOutput}
	tx, _ := transaction.NewIssueAssetTransaction(outputs)
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)
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

func makeTransferTransaction(signer *account.Account, programHashStr, assetHashStr string, value Fixed64) (string, error) {
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	myProgramHashStr := ToHexString(signer.ProgramHash.ToArray())

	resp, err := httpjsonrpc.Call(Address(), "getunspendoutput", 0, []interface{}{myProgramHashStr, assetHashStr})
	if err != nil {
		fmt.Println("HTTP JSON call failed")
		return "", err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return "", err
	}

	inputs := []*transaction.UTXOTxInput{}
	outputs := []*transaction.TxOutput{}
	transferTxOutput := &transaction.TxOutput{
		AssetID:     assetHash,
		Value:       value,
		ProgramHash: programHash,
	}
	outputs = append(outputs, transferTxOutput)

	unspend := r["result"].(map[string]interface{})
	expected := transferTxOutput.Value
	for k, v := range unspend {
		h := k[0:REFERTXHASHLEN]
		i := k[REFERTXHASHLEN+1:]
		b, _ := hex.DecodeString(h)
		var referHash Uint256
		referHash.Deserialize(bytes.NewReader(b))
		referIndex, _ := strconv.Atoi(i)

		out := v.(map[string]interface{})
		value := Fixed64(out["Value"].(float64))
		if value == expected {
			transferUTXOInput := &transaction.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			expected = 0
			inputs = append(inputs, transferUTXOInput)
			break
		} else if value > expected {
			transferUTXOInput := &transaction.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			inputs = append(inputs, transferUTXOInput)
			getChangeOutput := &transaction.TxOutput{
				AssetID:     assetHash,
				Value:       value - expected,
				ProgramHash: signer.ProgramHash,
			}
			expected = 0
			outputs = append(outputs, getChangeOutput)
			break
		} else if value < expected {
			transferUTXOInput := &transaction.UTXOTxInput{
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
		return "", errors.New("transfer failed, ammount is not enough")
	}
	tx, _ := transaction.NewTransferAssetTransaction(inputs, outputs)
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

func assetAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	reg := c.Bool("reg")
	issue := c.Bool("issue")
	transfer := c.Bool("transfer")
	if !reg && !issue && !transfer {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	wallet := openWallet(c.String("wallet"), WalletPassword(c.String("password")))
	admin, _ := wallet.GetDefaultAccount()
	value := c.Int64("value")
	if value == 0 {
		fmt.Println("invalid value [--value]")
		return nil
	}

	var txHex string
	var err error
	if reg {
		name := c.String("name")
		if name == "" {
			rbuf := make([]byte, RANDBYTELEN)
			rand.Read(rbuf)
			name = "DNA-" + ToHexString(rbuf)
		}
		issuer := admin
		description := "description"
		txHex, err = makeRegTransaction(admin, issuer, name, description, Fixed64(value))
	} else {
		asset := c.String("asset")
		to := c.String("to")
		if asset == "" || to == "" {
			fmt.Println("missing flag [--asset] or [--to]")
			return nil
		}
		if issue {
			txHex, err = makeIssueTransaction(admin, to, asset, Fixed64(value))
		} else if transfer {
			txHex, err = makeTransferTransaction(admin, to, asset, Fixed64(value))
		}
		if err != nil {
			fmt.Println(err)
			return nil
		}
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
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "asset ammount",
			},
		},
		Action: assetAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "asset")
			return cli.NewExitError("", 1)
		},
	}
}
