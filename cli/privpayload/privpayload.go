package privpayload

import (
	"DNA/account"
	. "DNA/cli/common"
	"DNA/core/contract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"DNA/core/transaction/payload"
	"DNA/crypto"
	"DNA/net/httpjsonrpc"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"

	. "github.com/bitly/go-simplejson"
	"github.com/urfave/cli"
	"strconv"
)

func makePrivacyTx(admin *account.Account, toPubkeyStr string, pload string) (string, error) {
	data := []byte(pload)
	toPk, _ := hex.DecodeString(toPubkeyStr)
	toPubkey, _ := crypto.DecodePoint(toPk)

	tx, _ := transaction.NewPrivacyPayloadTransaction(admin.PrivateKey, admin.PublicKey, toPubkey, payload.RawPayload, data)
	txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &txAttr)
	if err := signTransaction(admin, tx); err != nil {
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
	transactionContractContext := &contract.ContractContext{
		Data:       tx,
		Codes:      make([][]byte, 1),
		Parameters: make([][][]byte, 1),
	}

	if err := transactionContractContext.AddContract(transactionContract, signer.PubKey(), signature); err != nil {
		fmt.Println("AddContract failed")
		return err
	}
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return nil
}

func privpayloadAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	enc := c.Bool("enc")
	dec := c.Bool("dec")
	if !enc && !dec {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	wallet := account.Open(c.String("wallet"), WalletPassword(c.String("password")))
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", c.String("wallet"))
		os.Exit(1)
	}
	admin, _ := wallet.GetDefaultAccount()
	if enc {
		data := c.String("data")
		to := c.String("to")

		txHex, err := makePrivacyTx(admin, to, data)
		resp, err := httpjsonrpc.Call(Address(), "sendrawtransaction", 0, []interface{}{txHex})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		FormatOutput(resp)
	}

	if dec {
		txhash := c.String("txhash")
		resp, err := httpjsonrpc.Call(Address(), "getrawtransaction", 0, []interface{}{txhash})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		js, err := NewJson(resp)
		txType, _ := js.Get("result").Get("TxType").Int()
		if transaction.TransactionType(txType) != transaction.PrivacyPayload {
			return errors.New("txType error")
		}

		plDataStr, _ := js.Get("result").Get("Payload").Get("Payload").String()
		plData, _ := hex.DecodeString(plDataStr)

		enType, _ := js.Get("result").Get("Payload").Get("EncryptType").Int()
		switch payload.PayloadEncryptType(enType) {
		case payload.ECDH_AES256:
			enAttr, _ := js.Get("result").Get("Payload").Get("EncryptAttr").String()
			Attr, _ := hex.DecodeString(enAttr)
			bytesBuffer := bytes.NewBuffer(Attr)
			encryptAttr := new(payload.EcdhAes256)
			encryptAttr.Deserialize(bytesBuffer)

			if admin.PublicKey.X.Cmp(encryptAttr.ToPubkey.X) != 0 {
				fmt.Println("The wallet is wrong")
				return errors.New("The wallet is wrong")
			}
			privkey := admin.PrivateKey
			data, _ := encryptAttr.Decrypt(plData, privkey)

			//		encoding, _ := json.Marshal(map[string]string{"result": hex.EncodeToString(data)})
			encoding, _ := json.Marshal(map[string]string{"result": string(data)})
			FormatOutput(encoding)

		default:
			return errors.New("enType error")
		}
	}

	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "privacy",
		Usage:       "support encryption for payloads",
		Description: "With nodectl privacy, you could create privacy payload.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "enc, e",
				Usage: "create an privacy  payload",
			},
			cli.BoolFlag{
				Name:  "dec, d",
				Usage: "decrypt the privacy payload",
			},
			cli.StringFlag{
				Name:  "to",
				Usage: "payload to whom",
			},
			cli.StringFlag{
				Name:  "data",
				Usage: "data to be encrypted",
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
				Name:  "txhash, t",
				Usage: "hash of a transaction",
			},
		},
		Action: privpayloadAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "privacyPayload")
			return cli.NewExitError("", 1)
		},
	}
}
