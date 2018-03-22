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

package debug

//import (
//	"fmt"
//	"os"
//
//	. "github.com/Ontology/cli/common"
//	"github.com/Ontology/http/httpjsonrpc"
//
//	"github.com/urfave/cli"
//)
//
//func debugAction(c *cli.Context) (err error) {
//	if c.NumFlags() == 0 {
//		cli.ShowSubcommandHelp(c)
//		return nil
//	}
//	level := c.Int("level")
//	if level != -1 {
//		resp, err := jsonrpc.Call(Address(), "setdebuginfo", 0, []interface{}{level})
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err)
//			return err
//		}
//		FormatOutput(resp)
//	}
//	return nil
//}
//
//func NewCommand() *cli.Command {
//	return &cli.Command{Name: "debug",
//		Usage:       "blockchain node debugging",
//		Description: "With nodectl debug, you could debug blockchain node.",
//		ArgsUsage:   "[args]",
//		Flags: []cli.Flag{
//			cli.IntFlag{
//				Name:  "level, l",
//				Usage: "log level 0-6",
//				Value: -1,
//			},
//		},
//		Action: debugAction,
//		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
//			PrintError(c, err, "debug")
//			return cli.NewExitError("", 1)
//		},
//	}
//}
