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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	jrpc "github.com/ontio/ontology/http/base/rpc"
	"github.com/urfave/cli"
)

var (
	SettingCommand = cli.Command{
		Name:        "set",
		Usage:       "./ontology set debug/consensus [OPTION]",
		Category:    "Setting COMMANDS",
		Description: ``,
		Subcommands: []cli.Command{
			{
				Action:      utils.MigrateFlags(debugCommand),
				Name:        "debug",
				Usage:       "./ontology set debug [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), ContractFlags...),
				Category:    "SETTING COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(consensusCommand),
				Name:        "consensus",
				Usage:       "./ontology set consensus [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), ContractFlags...),
				Category:    "SETTING COMMANDS",
				Description: ``,
			},
		},
	}
)

func localRpcAddress() string {
	return "http://localhost:" + strconv.Itoa(config.Parameters.HttpJsonPort)
}

func debugCommand(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if client == nil {
		log.Fatal("Can't get local account.")
	}

	level := ctx.GlobalUint(utils.DebugLevelFlag.Name)
	resp, err := jrpc.Call(localRpcAddress(), "setdebuginfo", 0, []interface{}{level})
	if nil != err {
		return err
	}
	r := make(map[string]interface{})
	json.Unmarshal(resp, &r)
	fmt.Printf("%v\n", r)
	return nil
}

func consensusCommand(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if client == nil {
		log.Fatal("Can't get local account.")
	}

	on := ctx.GlobalUint(utils.ConsensusLevelFlag.Name)
	var resp []byte
	var err error
	switch on {
	case 1:
		resp, err = jrpc.Call(localRpcAddress(), "startconsensus", 0, []interface{}{on})
	case 0:
		resp, err = jrpc.Call(localRpcAddress(), "stopconsensus", 0, []interface{}{on})
	default:
		fmt.Println("Start:1; Stop:0; Pls enter valid value between 0 and 1.")
	}
	if nil != err {
		return err
	}
	r := make(map[string]interface{})
	json.Unmarshal(resp, &r)
	fmt.Printf("%v\n", r)
	return nil
}
