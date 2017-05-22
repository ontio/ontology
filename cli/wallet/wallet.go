package wallet

import (
	//"fmt"
	//	"os"

	. "DNA/cli/common"
	"DNA/client"
	"DNA/crypto"

	"github.com/urfave/cli"
)

func walletAction(c *cli.Context) (err error) {
	crypto.SetAlg(crypto.P256R1)
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	create := c.Bool("create")
	name := c.String("name")
	passwd := c.String("password")
	if create {
		if name != "" && passwd != "" {
			wallet := client.CreateClient(name, []byte(passwd))
			_ = wallet
		}
	} else {
		if name != "" && passwd != "" {
			wallet := client.OpenClient(name, []byte(passwd))
			_ = wallet
		}
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "wallet",
		Usage:       "user wallet operation",
		Description: "With nodectl wallet, you could do transaction.",
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
				Name:  "name, n",
				Usage: "wallet name",
				Value: "wallet.db3",
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: "dnapw",
			},
		},
		Action: walletAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "wallet")
			return cli.NewExitError("", 1)
		},
	}
}
