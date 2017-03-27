package info

import (
	"DNA/net/httpjsonrpc"
	"DNA/utility"
	"fmt"
	"os"
)

var usage = `show info about blockchain`

var flags = []string{"ip", "port", "rpcid", "height", "bestblockhash", "blockhash",
	"blockcount", "connections", "neighbor", "state"}

func main(args []string, p utility.Param) (err error) {
	var resp []byte
	var output [][]byte
	addr := utility.Address(p.Ip, p.Port)
	id := p.RPCID
	if height := p.Height; height >= 0 {
		resp, err = httpjsonrpc.Call(addr, "getblock", id, []interface{}{height, 1})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if hash := p.BlockHash; hash != "" {
		resp, err = httpjsonrpc.Call(addr, "getblock", id, []interface{}{hash, 1})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if p.BestBlockHash {
		resp, err = httpjsonrpc.Call(addr, "getbestblockhash", id, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if p.BlockCount {
		resp, err = httpjsonrpc.Call(addr, "getblockcount", id, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if p.ConnectionCount {
		resp, err = httpjsonrpc.Call(addr, "getconnectioncount", id, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if p.Neighbor {
		resp, err := httpjsonrpc.Call(addr, "getneighbor", id, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	if p.NodeState {
		resp, err := httpjsonrpc.Call(addr, "getnodestate", id, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		output = append(output, resp)
	}

	for _, v := range output {
		utility.FormatOutput(v)
	}

	return nil
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: main}
