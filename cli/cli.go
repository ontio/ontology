package cli

import (
	"os"
	"sort"

	. "DNA/cli/common"
	"DNA/cli/consensus"
	"DNA/cli/debug"
	"DNA/cli/info"
	"DNA/cli/test"

	"github.com/urfave/cli"
)

func init() {
	app := cli.NewApp()
	app.Name = "nodectl"
	app.Version = "1.0.1"
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
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
