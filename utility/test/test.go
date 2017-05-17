package test

import (
	"DNA/net/httpjsonrpc"
	"DNA/utility"
	"fmt"
	"os"
)

var usage = `run sample routines`

var flags = []string{"tx", "num"}

func testMain(args []string, p utility.Param) (err error) {
	if txType := p.Tx; txType != "" {
		resp, err := httpjsonrpc.Call(utility.Address(p.Ip, p.Port), "sendsampletransaction",
			p.RPCID, []interface{}{p.Tx, p.TxNum})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		utility.FormatOutput(resp)
	}
	return nil
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: testMain}
