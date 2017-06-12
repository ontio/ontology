package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"DNA/account"
	. "DNA/cli/common"
	. "DNA/common"
	"DNA/core/contract"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

func walletAction(c *cli.Context) (err error) {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	createWallet := c.Bool("create")
	list := c.Bool("list")
	name := c.String("name")
	passwd := c.String("password")
	if name != "" && passwd != "" {
		var wallet *account.ClientImpl
		if createWallet {
			wallet = account.Create(name, []byte(passwd))
		} else if list {
			wallet = account.Open(name, []byte(passwd))
		}
		if wallet == nil {
			fmt.Println("Failed to open wallet: ", name)
			os.Exit(1)
		}
		account, _ := wallet.GetDefaultAccount()
		pubKey := account.PubKey()
		signatureRedeemScript, _ := contract.CreateSignatureRedeemScript(pubKey)
		programHash, _ := ToCodeHash(signatureRedeemScript)
		encodedPubKey, _ := pubKey.EncodePoint(true)
		asset := c.String("asset")
		if asset == "" {
			fmt.Println("public key:   ", ToHexString(encodedPubKey))
			fmt.Println("program hash: ", ToHexString(programHash.ToArray()))
			fmt.Println("address:      ", programHash.ToAddress())
		}
		if list && asset != "" {
			var buffer bytes.Buffer
			_, err := programHash.Serialize(&buffer)
			if err != nil {
				return err
			}
			resp, err := httpjsonrpc.Call(Address(), "getunspendoutput", 0,
				[]interface{}{hex.EncodeToString(buffer.Bytes()), asset})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
			r := make(map[string]interface{})
			err = json.Unmarshal(resp, &r)
			if err != nil {
				fmt.Println("Unmarshal JSON failed")
				return err
			}
			switch r["result"].(type) {
			case map[string]interface{}:
				ammount := 0
				unspend := r["result"].(map[string]interface{})
				for _, v := range unspend {
					out := v.(map[string]interface{})
					ammount += int(out["Value"].(float64))
				}
				fmt.Println("Ammount: ", ammount)
			case string:
				fmt.Println(r["result"].(string))
				return nil
			}
		}
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "wallet",
		Usage:       "user wallet operation",
		Description: "With nodectl wallet, you could control your asset.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "create, c",
				Usage: "create wallet",
			},
			cli.BoolFlag{
				Name:  "list, l",
				Usage: "list wallet information",
			},
			cli.StringFlag{
				Name:  "asset, a",
				Usage: "asset uniq ID",
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "wallet name",
				Value: account.WalletFileName,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: account.DefaultPin,
			},
		},
		Action: walletAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "wallet")
			return cli.NewExitError("", 1)
		},
	}
}
