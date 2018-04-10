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

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/password"

	"github.com/urfave/cli"
)

var (
	Ip   string
	Port string
)

func NewIpFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "ip",
		Usage:       "node's ip address",
		Value:       "localhost",
		Destination: &Ip,
	}
}

func NewPortFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "port",
		Usage:       "node's RPC port",
		Value:       strconv.Itoa(config.Parameters.HttpLocalPort),
		Destination: &Port,
	}
}

func RpcAddress() string {
	address := "http://" + Ip + ":" + Port
	return address
}

func PrintError(c *cli.Context, err error, cmd string) {
	fmt.Println("Incorrect Usage:", err)
	fmt.Println("")
	cli.ShowCommandHelp(c, cmd)
}

func FormatOutput(o []byte) error {
	var out bytes.Buffer
	err := json.Indent(&out, o, "", "\t")
	if err != nil {
		return err
	}
	out.Write([]byte("\n"))
	_, err = out.WriteTo(os.Stdout)

	return err
}

// WalletPassword prompts user to input wallet password when password is not
// specified from command line
func WalletPassword(passwd string) []byte {
	if passwd == "" {
		tmppasswd, _ := password.GetPassword()
		return tmppasswd
	} else {
		return []byte(passwd)
	}
}
