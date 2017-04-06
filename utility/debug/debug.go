package debug

import (
	"DNA/net/httpjsonrpc"
	"DNA/utility"
	"fmt"
	"os"
)

var usage = `set debugging function`

var flags = []string{"level"}

func debugMain(args []string, p utility.Param) (err error) {
	var output [][]byte
	addr := utility.Address(p.Ip, p.Port)
	id := p.RPCID
	resp, err := httpjsonrpc.Call(addr, "setdebuginfo", id, []interface{}{p.DebugLevel})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	output = append(output, resp)
	for _, v := range output {
		utility.FormatOutput(v)
	}
	return nil
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: debugMain}
