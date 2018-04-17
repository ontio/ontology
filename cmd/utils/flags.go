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

package utils

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli"
)

type DirectoryString struct {
	Value string
}

type DirectoryFlag struct {
	Name  string
	Value DirectoryString
	Usage string
}

var (
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
	}

	EncryptTypeFlag = cli.StringFlag{
		Name: "encrypt",
		Usage: `assign encrypt type when use create wallet, just as:
						SHA224withECDSA, SHA256withECDSA,
						SHA384withECDSA, SHA512withECDSA,
						SHA3-224withECDSA, SHA3-256withECDSA,
						SHA3-384withECDSA, SHA3-512withECDSA,
						RIPEMD160withECDSA, SM3withSM2, SHA512withEdDSA`,
	}

	WalletAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "wallet address string(base58)",
	}

	WalletNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "wallet name",
	}

	WalletPwdFlag = cli.StringFlag{
		Name:  "pwd",
		Usage: "wallet password when used",
	}

	WalletUsedFlag = cli.StringFlag{
		Name:  "wallet",
		Usage: "which wallet will be used",
	}

	ConfigUsedFlag = cli.StringFlag{
		Name:  "config",
		Usage: "which config file will be used",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "enable rpc server? true or false",
	}

	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "enable websocket server? true or false",
	}

	//information cmd settings
	BHashInfoFlag = cli.StringFlag{
		Name:  "bhash",
		Usage: "block hash value",
	}

	BTrxInfoFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "transaction hash value",
	}

	HeightInfoFlag = cli.StringFlag{
		Name:  "height",
		Usage: "block height value",
	}

	//send raw transaction
	ContractAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "contract address that will be used when send raw transaction",
	}

	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "address which transfer from",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "address which transfer to",
	}
	TransactionValueFlag = cli.Int64Flag{
		Name:  "value",
		Usage: "transfer value",
	}
	UserPasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "used when transfer",
	}

	DebugLevelFlag = cli.UintFlag{
		Name:  "level",
		Usage: "debug level(0~6) will be set",
	}

	ConsensusLevelFlag = cli.UintFlag{
		Name:  "on",
		Usage: "consensus turn on/off",
	}

	//contract deploy
	ContractVmTypeFlag = cli.UintFlag{
		Name:  "type",
		Usage: "contract type ,value: NEOVM | NATIVE | SWAM",
	}

	ContractStorageFlag = cli.BoolFlag{
		Name:  "store",
		Usage: "does this contract will be stored, value: true or false",
	}

	ContractCodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "directory of smart contract that will be deployed",
	}

	ContractNameFlag = cli.StringFlag{
		Name:  "cname",
		Usage: "contract name that will be deployed",
	}

	ContractVersionFlag = cli.StringFlag{
		Name:  "cversion",
		Usage: "contract version which will be deployed",
	}

	ContractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "owner of deployed smart contract",
	}

	ContractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "owner email when deploy a smart contract",
	}

	ContractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "contract description when deploy one",
	}

	ContractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "contract parameter needed when invoded",
	}
	//contract invoke
)

func (self *DirectoryFlag) Set(value string) {
	self.Value.Value = value
}

// Expands a file path
// 1. replace tilde with users home dir
// 2. expands embedded environment variables
// 3. cleans the path, e.g. /a/b/../c -> /a/c
// Note, it has limitations, e.g. ~someuser/tmp will not be expanded
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := os.Getenv("HOME"); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func (self *DirectoryString) Set(value string) error {
	self.Value = expandPath(value)
	return nil
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := os.Getenv("HOME")
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Ontology")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Ontology")
		} else {
			return filepath.Join(home, ".ontology")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

// MigrateFlags sets the global flag from a local flag when it's set.
// This is a temporary function used for migrating old command/flags to the
// new format.
//
// e.g. geth account new --keystore /tmp/mykeystore --lightkdf
//
// is equivalent after calling this method with:
//
// geth --keystore /tmp/mykeystore --lightkdf account new
//
// This allows the use of the existing configuration functionality.
// When all flags are migrated this function can be removed and the existing
// configuration functionality must be changed that is uses local flags
func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

func absolutePath(Datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(Datadir, filename)
}

func eachName(longName string, fn func(string)) {
	parts := strings.Split(longName, ",")
	for _, name := range parts {
		name = strings.Trim(name, " ")
		fn(name)
	}
}

func (self DirectoryFlag) GetName() string {
	return self.Name
}

func (self *DirectoryString) String() string {
	return self.Value
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

func prefixedNames(fullName string) (prefixed string) {
	parts := strings.Split(fullName, ",")
	for i, name := range parts {
		name = strings.Trim(name, " ")
		prefixed += prefixFor(name) + name
		if i < len(parts)-1 {
			prefixed += ", "
		}
	}
	return
}

func (self DirectoryFlag) String() string {
	fmtString := "%s %v\t%v"
	if len(self.Value.Value) > 0 {
		fmtString = "%s \"%v\"\t%v"
	}
	return fmt.Sprintf(fmtString, prefixedNames(self.Name), self.Value.Value, self.Usage)
}

// called by cli library, grabs variable from environment (if in env)
// and adds variable to flag set for parsing.
func (self DirectoryFlag) Apply(set *flag.FlagSet) {
	eachName(self.Name, func(name string) {
		set.Var(&self.Value, self.Name, self.Usage)
	})
}
