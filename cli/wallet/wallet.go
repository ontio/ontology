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
	"DNA/common/password"
	"DNA/core/contract"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

func walletAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	// wallet name is wallet.dat by default
	name := c.String("name")
	create := c.Bool("create")
	list := c.Bool("list")
	passwd := c.String("password")
	if name == "" {
		fmt.Println("Invalid wallet name.")
		os.Exit(1)
	}
	if FileExisted(name) && create {
		fmt.Printf("CAUTION: '%s' already exists!\n", name)
		os.Exit(1)
	}
	// need to input password when password is not specified from command line
	if passwd == "" {
		var err error
		var tmppasswd []byte
		if create {
			tmppasswd, err = password.GetConfirmedPassword()
		} else {
			tmppasswd, err = password.GetPassword()
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		passwd = string(tmppasswd)
	}
	var wallet *account.ClientImpl
	if create {
		wallet = account.Create(name, []byte(passwd))
	} else {
		// list wallet or change wallet password
		wallet = account.Open(name, []byte(passwd))
	}
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}
	fmt.Printf("Wallet File: '%s'\n", name)
	if c.Bool("changepassword") {
		fmt.Println("# input new password #")
		newPassword, _ := password.GetConfirmedPassword()
		if ok := wallet.ChangePassword([]byte(passwd), newPassword); !ok {
			fmt.Println("error: failed to change password")
			os.Exit(1)
		}
		fmt.Println("password changed")
		return nil
	}
	account, _ := wallet.GetDefaultAccount()
	pubKey := account.PubKey()
	signatureRedeemScript, _ := contract.CreateSignatureRedeemScript(pubKey)
	programHash, _ := ToCodeHash(signatureRedeemScript)
	encodedPubKey, _ := pubKey.EncodePoint(true)
	address, _ := programHash.ToAddress()
	fmt.Println("public key:   ", ToHexString(encodedPubKey))
	fmt.Println("program hash: ", ToHexString(programHash.ToArray()))
	fmt.Println("address:      ", address)
	asset := c.String("asset")
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
			cli.BoolFlag{
				Name:  "changepassword",
				Usage: "change wallet password",
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
			},
		},
		Action: walletAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "wallet")
			return cli.NewExitError("", 1)
		},
	}
}
