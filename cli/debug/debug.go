package debug

import (
	"fmt"
	"os"

	. "DNA/cli/common"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

func debugAction(c *cli.Context) (err error) {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	level := c.Int("level")
	if level != -1 {
		resp, err := httpjsonrpc.Call(Address(), "setdebuginfo", 0, []interface{}{level})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		FormatOutput(resp)
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{Name: "debug",
		Usage:       "blockchain node debugging",
		Description: "With nodectl debug, you could debug blockchain node.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "level, l",
				Usage: "log level 0-6",
				Value: -1,
			},
		},
		Action: debugAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "debug")
			return cli.NewExitError("", 1)
		},
	}
}
