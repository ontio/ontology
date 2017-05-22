package test

import (
	"fmt"
	"os"

	. "DNA/cli/common"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

func testAction(c *cli.Context) (err error) {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	txnType := c.String("tx")
	txnNum := c.Int("num")
	if txnType != "" {
		resp, err := httpjsonrpc.Call(Address(), "sendsampletransaction", 0, []interface{}{txnType, txnNum})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		FormatOutput(resp)
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "run test routine",
		Description: "With nodectl test, you could run simple tests.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "tx, t",
				Usage: "send sample transaction",
				Value: "perf",
			},
			cli.IntFlag{
				Name:  "num, n",
				Usage: "sample transaction numbers",
				Value: 1,
			},
		},
		Action: testAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "test")
			return cli.NewExitError("", 1)
		},
	}
}
