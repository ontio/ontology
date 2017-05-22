package consensus

import (
	"fmt"
	"os"

	. "DNA/cli/common"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

func consensusAction(c *cli.Context) (err error) {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	var resp []byte
	if c.Bool("start") {
		resp, err = httpjsonrpc.Call(Address(), "startconsensus", 0, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		FormatOutput(resp)
	}
	if c.Bool("stop") {
		resp, err = httpjsonrpc.Call(Address(), "stopconsensus", 0, []interface{}{})
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
		Name:        "consensus",
		Usage:       "switch of consensue function",
		Description: "With nodectl consensus, you could start or stop consensus for a node.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "start",
				Usage: "start consensue",
			},
			cli.BoolFlag{
				Name:  "stop",
				Usage: "stop consensue",
			},
		},
		Action: consensusAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "consensus")
			return cli.NewExitError("", 1)
		},
	}
}
