package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
	"strings"
)

var MultiSigCommand = cli.Command{
	Name:  "multisig",
	Usage: "Multi signature operation. ",
	Subcommands: []cli.Command{
		{
			Name:        "genaddr",
			Usage:       "Generate multi signature address",
			Description: "Generate multi signature address.",
			Action:      genMultiAddress,
			Flags: []cli.Flag{
				utils.AccountMultiMFlag,
				utils.AccountMultiPubKeyFlag,
			},
		},
		{
			Name:        "sigtx",
			Usage:       "Sign to multi-signature transaction",
			ArgsUsage:   "<rawtx>",
			Description: "Sign to multi-signature transaction.",
			Action:      sigToMultiTx,
			Flags: []cli.Flag{
				utils.WalletFileFlag,
				utils.AccountMultiMFlag,
				utils.AccountMultiPubKeyFlag,
				utils.AccountAddressFlag,
				utils.AccountPassFlag,
			},
		},
	},

	Description: "Multi signature operation. Such as generate multi signature address, multi sign to transaction.",
}

func genMultiAddress(ctx *cli.Context) error {
	pkstr := strings.TrimSpace(strings.Trim(ctx.String(utils.GetFlagName(utils.AccountMultiPubKeyFlag)), ","))
	m := ctx.Uint(utils.GetFlagName(utils.AccountMultiMFlag))
	if pkstr == "" || m == 0 {
		PrintErrorMsg("Missing argument. %s or %s expected.",
			utils.GetFlagName(utils.AccountMultiMFlag),
			utils.GetFlagName(utils.AccountMultiPubKeyFlag))
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pks := strings.Split(pkstr, ",")
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for _, pk := range pks {
		pk := strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		data, err := hex.DecodeString(pk)
		pubKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return fmt.Errorf("invalid pub key:%s", pk)
		}
		pubKeys = append(pubKeys, pubKey)
	}
	pkSize := len(pubKeys)
	if !(1 <= m && int(m) <= pkSize && pkSize > 1 && pkSize <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		PrintErrorMsg("Invalid argument. %s must > 1 and <= %d, and m must > 0 and < number of pub key.",
			utils.GetFlagName(utils.AccountMultiPubKeyFlag),
			constants.MULTI_SIG_MAX_PUBKEY_SIZE)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	addr, err := types.AddressFromMultiPubKeys(pubKeys, int(m))
	if err != nil {
		return err
	}

	PrintInfoMsg("Pub key list:")
	for i, pubKey := range pubKeys {
		addr := types.AddressFromPubKey(pubKey)
		PrintInfoMsg("Index %d Address:%s PubKey:%x ", i+1, addr.ToBase58(), keypair.SerializePublicKey(pubKey))
	}
	PrintInfoMsg("\nMultiSigAddress:%s", addr.ToBase58())
	return nil
}

func sigToMultiTx(ctx *cli.Context) error {
	pkstr := strings.TrimSpace(strings.Trim(ctx.String(utils.GetFlagName(utils.AccountMultiPubKeyFlag)), ","))
	m := ctx.Uint(utils.GetFlagName(utils.AccountMultiMFlag))
	if pkstr == "" || m == 0 {
		PrintErrorMsg("Missing argument. %s or %s expected.",
			utils.GetFlagName(utils.AccountMultiMFlag),
			utils.GetFlagName(utils.AccountMultiPubKeyFlag))
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pks := strings.Split(pkstr, ",")
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for _, pk := range pks {
		pk := strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		data, err := hex.DecodeString(pk)
		pubKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return fmt.Errorf("invalid pub key:%s", pk)
		}
		pubKeys = append(pubKeys, pubKey)
	}
	pkSize := len(pubKeys)
	if !(1 <= m && int(m) <= pkSize && pkSize > 1 && pkSize <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		PrintErrorMsg("Invalid argument. %s must > 1 and <= %d, and m must > 0 and < number of pub key.",
			utils.GetFlagName(utils.AccountMultiPubKeyFlag),
			constants.MULTI_SIG_MAX_PUBKEY_SIZE)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	multiAddr, err := types.AddressFromMultiPubKeys(pubKeys, int(m))
	if err != nil {
		return err
	}

	PrintInfoMsg("Pub key list:")
	for i, pubKey := range pubKeys {
		addr := types.AddressFromPubKey(pubKey)
		PrintInfoMsg("  Index %d Pubkey:%x Address:%s", i+1, keypair.SerializePublicKey(pubKey), addr.ToBase58())
	}
	PrintInfoMsg("  MultiSigAddress:%s", multiAddr.ToBase58())

	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing %s flag.", utils.GetFlagName(utils.RawTransactionFlag))
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	rawTx := ctx.Args().First()
	txData, err := hex.DecodeString(rawTx)
	if err != nil {
		return fmt.Errorf("RawTx hex decode error:%s", err)
	}
	tx, err := types.TransactionFromRawBytes(txData)
	if err != nil {
		return fmt.Errorf("TransactionFromRawBytes error:%s", err)
	}

	mutTx, err := tx.IntoMutable()
	if err != nil {
		return fmt.Errorf("IntoMutable error:%s", err)
	}

	acc, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}
	err = utils.MultiSigTransaction(mutTx, uint16(m), pubKeys, acc)
	if err != nil {
		return fmt.Errorf("MultiSigTransaction error:%s", err)
	}

	tx, err = mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	err = tx.Serialization(&sink)
	if err != nil {
		return fmt.Errorf("tx serialization error:%s", err)
	}

	PrintInfoMsg("RawTx after signed:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}
