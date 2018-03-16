package transfer

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"github.com/Ontology/http/base/rpc"
	. "github.com/Ontology/cli/common"
	cutils "github.com/Ontology/core/utils"
	vmtypes "github.com/Ontology/vm/types"
	ctypes "github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/service/native/states"
	"github.com/Ontology/crypto"
	"math/big"
	"github.com/Ontology/common"
	"bytes"
	"github.com/Ontology/account"
	"encoding/json"
)

func transferAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	// wallet name is wallet.dat by default
	contract := c.String("contract")
	if contract == "" {
		fmt.Println("Invalid contract address.")
		os.Exit(1)
	}
	ct, _ := common.HexToBytes(contract)
	ctu, _ := common.Uint160ParseFromBytes(ct)
	from := c.String("from")
	if from == "" {
		fmt.Println("Invalid sender address.")
		os.Exit(1)
	}
	f, _ := common.HexToBytes(from)
	fu, _ := common.Uint160ParseFromBytes(f)
	to := c.String("to")
	if to == "" {
		fmt.Println("Invalid revicer address.")
		os.Exit(1)
	}
	t, _ := common.HexToBytes(from)
	tu, _ := common.Uint160ParseFromBytes(t)
	value := c.Int64("value")
	if value <= 0 {
		fmt.Println("Invalid ont amount.")
		os.Exit(1)
	}

	var sts []*states.State
	sts = append(sts, &states.State{
		From: fu,
		To: tu,
		Value: big.NewInt(value),
	})
	transfers := new(states.Transfers)
	transfers.Params = append(transfers.Params, &states.TokenTransfer{
		Contract: ctu,
		States: sts,
	})

	bf := new(bytes.Buffer)
	transfers.Deserialize(bf)

	tx := cutils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.NativeVM,
		Code: bf.Bytes(),
	})

	resp, err := rpc.Call(Address(), "sendrawtransaction", 0,
		[]interface{}{tx})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return err
	}
	fmt.Println("result", r["result"])
	switch r["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(r["result"].(string))
		return nil
	}
	return nil
}

func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
	signature, err := crypto.Sign(signer.PrivKey(), tx.GetMessage())
	if err != nil {
		return err
	}
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []*crypto.PubKey{signer.PublicKey},
		M: 1,
		SigData: [][]byte{signature},
	})
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "transfer",
		Usage:       "user ont transfer",
		Description: "With nodectl transfer, you could transfer ont.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "contract, c",
				Usage: "contract address",
			},
			cli.BoolFlag{
				Name:  "from, f",
				Usage: "sender address",
			},
			cli.BoolFlag{
				Name:  "to, t",
				Usage: "revicer address",
			},
			cli.StringFlag{
				Name:  "value, v",
				Usage: "ont amount",
			},
		},
		Action: transferAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
}

