package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Arith int

type Args struct {
	A, B int
}

type Reply struct {
	C int
}
type Result int

type GetBestBlockHashResp struct {
	Id     interface{} `json:"id"`
	Result Reply       `json:"result"`
	Error  interface{} `json:"error"`
}

func (t *Arith) Multiply(args *Args, result *Result) error {
	log.Printf("Multiplying %d with %d\n", args.A, args.B)
	print("test the multiply\n")
	*result = Result(args.A * args.B)
	return nil
}

// private JObject InternalCall(string method, JArray _params)
//         {
//             switch (method)
//             {
//                 case "getbestblockhash":
//                     return Blockchain.Default.CurrentBlockHash.ToString();
//                 case "getblock":
//                     {
//                         Block block;
//                         if (_params[0] is JNumber) {
//                             uint index = (uint)_params[0].AsNumber();
//                             block = Blockchain.Default.GetBlock(index);
//                         } else {
//                             UInt256 hash = UInt256.Parse(_params[0].AsString());
//                             block = Blockchain.Default.GetBlock(hash);
//                         }
//                         if (block == null)
//                             throw new RpcException(-100, "Unknown block");
//                         bool verbose = _params.Count >= 2 && _params[1].AsBooleanOrDefault(false);
//                         if (verbose)
//                             return block.ToJson();
// 		    }
// 	    }
// 	}
// }

// func Process(HttpListenerContext context)
// {
// 	request = JObject.Parse(reader)
// 	response = ProcessRequest(request)
// 	result = InternalCall(request["method"].AsString(), (JArray)request["params"]);
// }

func (t *Arith) getbestblockhash(args *Args, result *GetBestBlockHashResp) error {
	log.Printf("Multiplying with\n")
	print("test")
	//*result = Result(args.A * args.B)
	return nil
}

func startServer() {
	//arith := new(Arith)
	server := rpc.NewServer()
	//server.Register(arith)

	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	l, e := net.Listen("tcp", ":10333")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

func main() {
	go startServer()
	conn, err := net.Dial("tcp", "localhost:10333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	args := &Args{7, 8}

	var reply Reply
	c := jsonrpc.NewClient(conn)
	//c := rpc.NewClient(conn)
	dec := json.NewDecoder(conn)

	for i := 0; i < 1; i++ {
		err = c.Call("Arith.Multiply", args, &reply)
		//err = c.Call("Arith.getbestblockhash", args, &reply)
		fmt.Fprint(conn, `{"jsonrpc": "2.0", "method": "getbestblockhash", "params": [], "id": 2`)
		var resp GetBestBlockHashResp
		err := dec.Decode(&resp)
		if err != nil {
			log.Fatal("Decode: %s:", err)
		}
		fmt.Printf("Get best block hash resp: %d %d\n", resp.Id, resp.Result)
		print("test the end\n")
	}
}
