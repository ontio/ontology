package cli

import (
	"math/rand"
	"os"
	"sort"
	"time"

	"DNA/cli/asset"
	. "DNA/cli/common"
	"DNA/cli/consensus"
	"DNA/cli/debug"
	"DNA/cli/info"
	"DNA/cli/test"
	"DNA/cli/wallet"
	"DNA/common/log"
	"DNA/crypto"

	"github.com/urfave/cli"
	"DNA/common/config"
)

var Version string

func init() {
	var path string = "./Log/"
	log.CreatePrintLog(path)
	crypto.SetAlg(config.Parameters.EncryptAlg)
	//seed transaction nonce
	rand.Seed(time.Now().UnixNano())

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
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Run(os.Args)
}
