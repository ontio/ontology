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
	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
	"io"
	"sort"
)

// AppHelpTemplate is the test template for the default, global app help topic.
var (
	AppHelpTemplate = `NAME:
   {{.App.Name}} - {{.App.Usage}}

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
	{{range .App.Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
	{{end}}{{end}}{{if .FlagGroups}}
{{range .FlagGroups}}{{.Name}} OPTIONS:
	{{range .Flags}}{{.}}
	{{end}}
{{end}}{{end}}{{if .App.Copyright }}COPYRIGHT:
	{{.App.Copyright}}
{{end}}
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
{{range $categorized.Flags}}{{"\t"}}{{.}}
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
			utils.DisableEventLogFlag,
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
		},
	},
	{
		Name: "CLI",
		Flags: []cli.Flag{
			utils.CliEnableRpcFlag,
			utils.CliRpcPortFlag,
		},
	},
	{
		Name: "CONSENSUS",
		Flags: []cli.Flag{
			utils.DisableConsensusFlag,
			utils.MaxTxInBlockFlag,
		},
	},
	{
		Name: "P2P NODE",
		Flags: []cli.Flag{
			utils.NodePortFlag,
			utils.DualPortSupportFlag,
			utils.ConsensusPortFlag,
		},
	},
	{
		Name: "RPC",
		Flags: []cli.Flag{
			utils.RPCEnabledFlag,
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
		Name: "MISC",
	},
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
		for _, flg := range category.Flags {
			if flg.GetName() == flag.GetName() {
				return category.Name
			}
		}
	}
	return "MISC"
}

type cusHelpData struct {
	App        interface{}
	FlagGroups []flagGroup
}

func init() {
	//Using customize AppHelpTemplate
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate

	oriHelpPrinter := cli.HelpPrinter
	cusHelpPrinter := func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == AppHelpTemplate {
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
				App:        data,
				FlagGroups: sorted,
			}
			oriHelpPrinter(w, tmpl, cusData)
		} else if tmpl == CommandHelpTemplate {
			categorized := make(map[string][]cli.Flag)
			for _, flag := range data.(cli.Command).Flags {
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
			oriHelpPrinter(w, tmpl, map[string]interface{}{
				"cmd":              data,
				"categorizedFlags": sorted,
			})
		} else {
			oriHelpPrinter(w, tmpl, data)
		}
	}

	//Override the default global app help printer
	cli.HelpPrinter = cusHelpPrinter
}
