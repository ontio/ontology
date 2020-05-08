/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
)

// AppHelpTemplate is the test template for the default, global app help topic.
var (
	AppHelpTemplate = `NAME:
  {{.App.Name}} - {{.App.Usage}}

	Ontology CLI is an Ontology node command line Client for starting and managing Ontology nodes,
	managing user wallets, sending transactions, deploying and invoking contract, and so on.

USAGE:
  {{.App.HelpName}} [options]{{if .App.Commands}} command [command options] {{end}}{{if .App.ArgsUsage}}{{.App.ArgsUsage}}{{else}}[arguments...]{{end}}
  {{if .App.Version}}
VERSION:
  {{.App.Version}}
  {{end}}{{if len .App.Authors}}
AUTHOR(S):
  {{range .App.Authors}}{{ . }}{{end}}
  {{end}}{{if .App.Commands}}
COMMANDS:
  {{range .App.Commands}}{{join .Names ", "}}{{ "  " }}{{.Usage}}
  {{end}}{{end}}{{if .FlagGroups}}
{{range .FlagGroups}}{{.Name}} OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}
{{end}}{{end}}{{if .App.Copyright }}COPYRIGHT: 
  {{.App.Copyright}}
{{end}}
`
	SubcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

COMMANDS:
  {{range .Commands}}{{join .Names ", "}}{{ "  " }}{{.Usage}}
  {{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	CommandHelpTemplate = `
USAGE:
  {{if .cmd.UsageText}}{{.cmd.UsageText}}{{else}}{{.cmd.HelpName}}{{if .cmd.VisibleFlags}} [command options]{{end}} {{if .cmd.ArgsUsage}}{{.cmd.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .cmd.Description}}

DESCRIPTION:
  {{.cmd.Description}}
  {{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
  {{range .cmd.Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"  "}}{{.}}
{{end}}
{{end}}{{end}}`
)

//flagGroup is a collection of flags belonging to a single topic.
type flagGroup struct {
	Name  string
	Flags []cli.Flag
}

var AppHelpFlagGroups = []flagGroup{
	{
		Name: "ONTOLOGY",
		Flags: []cli.Flag{
			utils.ConfigFlag,
			utils.LogLevelFlag,
			utils.LogDirFlag,
			utils.DisableLogFileFlag,
			utils.DisableEventLogFlag,
			utils.DataDirFlag,
			utils.WasmVerifyMethodFlag,
		},
	},
	{
		Name: "ACCOUNT",
		Flags: []cli.Flag{
			utils.WalletFileFlag,
			utils.AccountAddressFlag,
			utils.AccountPassFlag,
			utils.AccountDefaultFlag,
			utils.AccountKeylenFlag,
			utils.AccountSetDefaultFlag,
			utils.AccountSigSchemeFlag,
			utils.AccountTypeFlag,
			utils.AccountVerboseFlag,
			utils.AccountLabelFlag,
			utils.AccountQuantityFlag,
			utils.AccountChangePasswdFlag,
			utils.AccountSourceFileFlag,
			utils.AccountWIFFlag,
			utils.AccountLowSecurityFlag,
			utils.AccountMultiMFlag,
			utils.AccountMultiPubKeyFlag,
			utils.IdentityFlag,
		},
	},
	{
		Name: "CONSENSUS",
		Flags: []cli.Flag{
			utils.EnableConsensusFlag,
			utils.MaxTxInBlockFlag,
		},
	},
	{
		Name: "TXPOOL",
		Flags: []cli.Flag{
			utils.GasPriceFlag,
			utils.GasLimitFlag,
			utils.TxpoolPreExecDisableFlag,
			utils.DisableSyncVerifyTxFlag,
			utils.DisableBroadcastNetTxFlag,
		},
	},
	{
		Name: "P2P NODE",
		Flags: []cli.Flag{
			utils.ReservedPeersOnlyFlag,
			utils.ReservedPeersFileFlag,
			utils.NetworkIdFlag,
			utils.NodePortFlag,
			utils.HttpInfoPortFlag,
			utils.MaxConnInBoundFlag,
			utils.MaxConnOutBoundFlag,
			utils.MaxConnInBoundForSingleIPFlag,
		},
	},
	{
		Name: "RPC",
		Flags: []cli.Flag{
			utils.RPCDisabledFlag,
			utils.RPCPortFlag,
			utils.RPCLocalEnableFlag,
			utils.RPCLocalProtFlag,
		},
	},
	{
		Name: "RESTFUL",
		Flags: []cli.Flag{
			utils.RestfulEnableFlag,
			utils.RestfulPortFlag,
			utils.RestfulMaxConnsFlag,
		},
	},
	{
		Name: "WEB SOCKET",
		Flags: []cli.Flag{
			utils.WsEnabledFlag,
			utils.WsPortFlag,
		},
	},
	{
		Name: "TEST MODE",
		Flags: []cli.Flag{
			utils.EnableTestModeFlag,
			utils.TestModeGenBlockTimeFlag,
		},
	},
	{
		Name: "CONTRACT",
		Flags: []cli.Flag{
			utils.ContractPrepareDeployFlag,
			utils.ContractAddrFlag,
			utils.ContractAuthorFlag,
			utils.ContractCodeFileFlag,
			utils.ContractDescFlag,
			utils.ContractEmailFlag,
			utils.ContractNameFlag,
			utils.ContractVersionFlag,
			utils.ContractVmTypeFlag,
			utils.ContractPrepareInvokeFlag,
			utils.ContractParamsFlag,
			utils.ContractReturnTypeFlag,
		},
	},
	{
		Name: "TRANSACTION",
		Flags: []cli.Flag{
			utils.TransactionGasLimitFlag,
			utils.TransactionGasPriceFlag,
			utils.TransactionAssetFlag,
			utils.TransactionFromFlag,
			utils.TransactionToFlag,
			utils.TransactionAmountFlag,
			utils.TransactionHashFlag,
			utils.TransferFromSenderFlag,
			utils.ApproveAssetFlag,
			utils.ApproveAssetFromFlag,
			utils.ApproveAssetToFlag,
			utils.ApproveAmountFlag,
			utils.SendTxFlag,
			utils.ForceSendTxFlag,
			utils.TransactionPayerFlag,
			utils.PrepareExecTransactionFlag,
			utils.TransferFromAmountFlag,
			utils.WithdrawONGReceiveAccountFlag,
			utils.WithdrawONGAmountFlag,
		},
	},
	{
		Name: "Approve",
		Flags: []cli.Flag{
			utils.ApproveAssetFromFlag,
			utils.ApproveAssetToFlag,
		},
	},
	{
		Name: "EXPORT",
		Flags: []cli.Flag{
			utils.ExportFileFlag,
			utils.ExportSpeedFlag,
			utils.ExportStartHeightFlag,
			utils.ExportEndHeightFlag,
		},
	},
	{
		Name: "IMPORT",
		Flags: []cli.Flag{
			utils.ImportFileFlag,
			utils.ImportEndHeightFlag,
		},
	},
	{
		Name: "MISC",
	},
}

func init() {
	//Using customize AppHelpTemplate
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate
	//Override the default global app help printer
	cli.HelpPrinter = cusHelpPrinter
}

// byCategory sorts flagGroup by Name in in the order of AppHelpFlagGroups.
type byCategory []flagGroup

func (a byCategory) Len() int      { return len(a) }
func (a byCategory) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCategory) Less(i, j int) bool {
	iCat, jCat := a[i].Name, a[j].Name
	iIdx, jIdx := len(AppHelpFlagGroups), len(AppHelpFlagGroups) // ensure non categorized flags come last

	for i, group := range AppHelpFlagGroups {
		if iCat == group.Name {
			iIdx = i
		}
		if jCat == group.Name {
			jIdx = i
		}
	}

	return iIdx < jIdx
}

func flagCategory(flag cli.Flag) string {
	for _, category := range AppHelpFlagGroups {
		for _, catFlag := range category.Flags {
			if catFlag.GetName() == flag.GetName() {
				catFlagString := strings.Replace(catFlag.String(), " ", "", -1)
				catFlagString = strings.Replace(catFlagString, "\t", "", -1)
				flagString := strings.Replace(flag.String(), " ", "", -1)
				if catFlagString == flagString {
					return category.Name
				}
			}
		}
	}
	return "MISC"
}

type cusHelpData struct {
	App        interface{}
	FlagGroups []flagGroup
}

type FmtFlag struct {
	name  string
	usage string
}

func (this *FmtFlag) GetName() string {
	return this.name
}

func (this *FmtFlag) String() string {
	return this.usage
}

func (this *FmtFlag) Apply(*flag.FlagSet) {}

func formatCommand(cmds []cli.Command) []cli.Command {
	maxWidth := 0
	for _, cmd := range cmds {
		if len(cmd.Name) > maxWidth {
			maxWidth = len(cmd.Name)
		}
	}
	formatter := "%-" + fmt.Sprintf("%d", maxWidth) + "s"
	newCmds := make([]cli.Command, 0, len(cmds))
	for _, cmd := range cmds {
		name := cmd.Name
		if len(cmd.Aliases) != 0 {
			for _, aliase := range cmd.Aliases {
				name += ", " + aliase
			}
			cmd.Aliases = nil
		}
		cmd.Name = fmt.Sprintf(formatter, name)
		newCmds = append(newCmds, cmd)
	}
	return newCmds
}

func formatFlags(flags []cli.Flag) []cli.Flag {
	maxWidth := 0
	fmtFlagStrs := make(map[string][]string)
	for _, flag := range flags {
		fs := strings.Split(flag.String(), "\t")
		if len(fs[0]) > maxWidth {
			maxWidth = len(fs[0])
		}
		fmtFlagStrs[flag.GetName()] = fs
	}
	formatter := "%-" + fmt.Sprintf("%d", maxWidth) + "s   %s"
	fmtFlags := make([]cli.Flag, 0, len(fmtFlagStrs))
	for _, flag := range flags {
		flagStrs := fmtFlagStrs[flag.GetName()]

		fmtFlags = append(fmtFlags, &FmtFlag{
			name:  flag.GetName(),
			usage: fmt.Sprintf(formatter, flagStrs[0], flagStrs[1]),
		})
	}
	return fmtFlags
}

func cusHelpPrinter(w io.Writer, tmpl string, data interface{}) {
	if tmpl == AppHelpTemplate {
		cliApp := data.(*cli.App)
		cliApp.Commands = formatCommand(cliApp.Commands)
		cliApp.Flags = formatFlags(cliApp.Flags)
		categorized := make(map[string][]cli.Flag)
		for _, flag := range data.(*cli.App).Flags {
			_, ok := categorized[flag.String()]
			if !ok {
				gName := flagCategory(flag)
				categorized[gName] = append(categorized[gName], flag)
			}
		}
		sorted := make([]flagGroup, 0, len(categorized))
		for cat, flgs := range categorized {
			sorted = append(sorted, flagGroup{cat, flgs})
		}
		sort.Sort(byCategory(sorted))
		cusData := &cusHelpData{
			App:        cliApp,
			FlagGroups: sorted,
		}
		data = cusData
	} else if tmpl == SubcommandHelpTemplate {
		cliApp := data.(*cli.App)
		cliApp.Commands = formatCommand(cliApp.Commands)
		data = cliApp
	} else if tmpl == CommandHelpTemplate {
		cliCmd := data.(cli.Command)
		cliCmd.Flags = formatFlags(cliCmd.Flags)
		categorized := make(map[string][]cli.Flag)
		for _, flag := range cliCmd.Flags {
			_, ok := categorized[flag.String()]
			if !ok {
				categorized[flagCategory(flag)] = append(categorized[flagCategory(flag)], flag)
			}
		}
		sorted := make([]flagGroup, 0, len(categorized))
		for cat, flgs := range categorized {
			sorted = append(sorted, flagGroup{cat, flgs})
		}
		sort.Sort(byCategory(sorted))
		data = map[string]interface{}{
			"cmd":              cliCmd,
			"categorizedFlags": sorted,
		}
	}

	funcMap := template.FuncMap{"join": strings.Join}
	t := template.Must(template.New("help").Funcs(funcMap).Parse(tmpl))
	err := t.Execute(w, data)
	if err != nil {
		// If the writer is closed, t.Execute will fail, and there's nothing we can do to recover.
		PrintErrorMsg("CLI TEMPLATE ERROR: %#v\n", err)
		return
	}
}

func PrintErrorMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[31m[ERROR] %s\033[0m\n", format) //Print error msg with red color
	fmt.Printf(format, a...)
}

func PrintWarnMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[33m[WARN] %s\033[0m\n", format) //Print error msg with yellow color
	fmt.Printf(format, a...)
}

func PrintInfoMsg(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

func PrintJsonData(data []byte) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "   ")
	if err != nil {
		PrintErrorMsg("json.Indent error:%s", err)
		return
	}
	PrintInfoMsg(out.String())
}

func PrintJsonObject(obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		PrintErrorMsg("json.Marshal error:%s", err)
		return
	}
	PrintJsonData(data)
}
