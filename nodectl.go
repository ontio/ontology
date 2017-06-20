package main

import (
	"os"
	"sort"

	_ "DNA/cli"
	"DNA/cli/asset"
	"DNA/cli/bookkeeper"
	. "DNA/cli/common"
	"DNA/cli/consensus"
	"DNA/cli/data"
	"DNA/cli/debug"
	"DNA/cli/info"
	"DNA/cli/privpayload"
	"DNA/cli/test"
	"DNA/cli/wallet"

	"github.com/urfave/cli"
)

var Version string

func main() {
	app := cli.NewApp()
	app.Name = "nodectl"
	app.Version = Version
	app.HelpName = "nodectl"
	app.Usage = "command line tool for DNA blockchain"
	app.UsageText = "nodectl [global options] command [command options] [args]"
	app.HideHelp = false
	app.HideVersion = false
	//global options
	app.Flags = []cli.Flag{
		NewIpFlag(),
		NewPortFlag(),
	}
	//commands
	app.Commands = []cli.Command{
		*consensus.NewCommand(),
		*debug.NewCommand(),
		*info.NewCommand(),
		*test.NewCommand(),
		*wallet.NewCommand(),
		*asset.NewCommand(),
		*privpayload.NewCommand(),
		*data.NewCommand(),
		*bookkeeper.NewCommand(),
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
