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

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/config"
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
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "SETTING COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(consensusCommand),
				Name:        "consensus",
				Usage:       "./ontology set consensus [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "SETTING COMMANDS",
				Description: ``,
			},
		},
	}
)

func localRpcAddress() string {
	return "http://127.0.0.1:" + strconv.Itoa(config.Parameters.HttpLocalPort)

}

func debugCommand(ctx *cli.Context) error {
	config.Init(ctx)
	level := ctx.GlobalUint(utils.DebugLevelFlag.Name)
	resp, err := jrpc.Call(localRpcAddress(), "setdebuginfo", 0, []interface{}{level})
	if nil != err {

	}
	r := make(map[string]interface{})
	json.Unmarshal(resp, &r)
	fmt.Printf("%v\n", r)
	return nil
}

func consensusCommand(ctx *cli.Context) error {
	config.Init(ctx)
	on := ctx.GlobalUint(utils.ConsensusLevelFlag.Name)
	var err error
	var resp []byte
	switch on {
	case 1:
		resp, err = jrpc.Call(localRpcAddress(), "startconsensus", 0, []interface{}{on})
	case 0:
		resp, err = jrpc.Call(localRpcAddress(), "stopconsensus", 0, []interface{}{on})
	}
	if nil != err {
		//log.Errorf("Set consensus error: %v", err)
	}
	r := make(map[string]interface{})
	json.Unmarshal(resp, &r)
	fmt.Printf("%v\n", r)
	return nil
}
