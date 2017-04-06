package utility

import (
	"DNA/config"
	"DNA/net/httpjsonrpc"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

var p Param
var flagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func init() {
	registerFlags(flagSet)
}

type Command struct {
	UsageText string
	Flags     []string
	Main      func(args []string, p Param) error
}

type Param struct {
	Ip              string // IP address
	Port            string // port number
	BestBlockHash   bool   // hash of current block
	Height          int64  // block height
	BlockHash       string // block hash
	BlockCount      bool   // block count
	ConnectionCount bool   // connected node number
	Neighbor        bool   // neighbor nodes
	Start           bool   // start service
	Stop            bool   // stop service
	NodeState       bool   // node state
	Tx              bool   // Transaction test case
	TxNum		int64  // count of transaction
	NoSign		bool   // transaction is not signed
	RPCID           int64  // RPC ID, use int64 by default
}

func registerFlags(f *flag.FlagSet) {
	f.StringVar(&p.Ip, "ip", "", "IP address")
	f.StringVar(&p.Port, "port", "", "port number")
	f.Int64Var(&p.RPCID, "rpcid", 0, "JSON-RPC ID")
	f.BoolVar(&p.BestBlockHash, "bestblockhash", false, "hash of current block")
	f.Int64Var(&p.Height, "height", -1, "height of blockchain")
	f.StringVar(&p.BlockHash, "blockhash", "", "block hash")
	f.BoolVar(&p.BlockCount, "blockcount", false, "block numbers")
	f.BoolVar(&p.ConnectionCount, "connections", false, "connection counnt")
	f.BoolVar(&p.Neighbor, "neighbor", false, "neighbor nodes information")
	f.BoolVar(&p.NodeState, "state", false, "node state")
	f.BoolVar(&p.Start, "start", false, "start service")
	f.BoolVar(&p.Tx, "tx", false, "send a sample transaction")
	f.Int64Var(&p.TxNum, "num", 1, "transaction number")
	f.BoolVar(&p.NoSign, "nosign", false, "send unsigned transaction")
	f.BoolVar(&p.Stop, "stop", false, "stop service")
}

func Start(cmds map[string]*Command) error {
	// Automatically called when -h present
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Available Commands:")
		for name := range cmds {
			fmt.Fprintf(os.Stderr, "  %-10s\t- %s\n", name, cmds[name].UsageText)
		}
	}
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "no subcommand is given\n")
		flag.Usage()
		return errors.New("no subcommand was given")
	}

	subCmdName := flag.Arg(0)
	subCmdArgs := flag.Args()[1:]
	subCmd, found := cmds[subCmdName]
	if !found {
		fmt.Fprintf(os.Stderr, "subcommand %s is not defined\n\n", subCmdName)
		flag.Usage()
		return errors.New("undefined subcommand")
	}

	// Automatically called when flagSet.Parse failed
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "\nAvailable Flags:")
		for _, name := range subCmd.Flags {
			if f := flagSet.Lookup(name); f != nil {
				fmt.Fprintf(os.Stderr, "  -%-15s\t%s\n", f.Name, f.Usage)
			}
		}
	}
	flagSet.Parse(subCmdArgs)
	if len(subCmdArgs) == 0 {
		fmt.Fprintf(os.Stderr, "no flag is given\n")
		flagSet.Usage()
		return errors.New("no flag for subcommand is given")
	}

	if err := subCmd.Main(subCmdArgs, p); err != nil {
		return err
	}

	return nil
}

func Address(ip, portnum string) string {
	// default address
	addr := "localhost"
	port := strconv.Itoa(config.Parameters.HttpLocalPort)
	if ip != "" {
		addr = ip
	}
	if portnum != "" {
		port = portnum
	}
	address := "http://" + addr + ":" + port + httpjsonrpc.LocalDir
	fmt.Printf("Connecting to %s ...\n", address)
	return address
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
