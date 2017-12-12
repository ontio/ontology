package main

import (
	"os"
	"sort"

	_ "github.com/Ontology/cli"
	"github.com/Ontology/cli/asset"
	"github.com/Ontology/cli/bookkeeper"
	. "github.com/Ontology/cli/common"
	"github.com/Ontology/cli/data"
	"github.com/Ontology/cli/debug"
	"github.com/Ontology/cli/info"
	"github.com/Ontology/cli/privpayload"
	"github.com/Ontology/cli/test"
	"github.com/Ontology/cli/wallet"

	"github.com/urfave/cli"
)

var Version string

func main() {
	app := cli.NewApp()
	app.Name = "nodectl"
	app.Version = Version
	app.HelpName = "nodectl"
	app.Usage = "command line tool for Ontology blockchain"
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
