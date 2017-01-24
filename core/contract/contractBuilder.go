package contract

import (
	"GoOnchain/crypto"
	"GoOnchain/vm"
	. "GoOnchain/common"
	pg "GoOnchain/core/contract/program"
	"math/big"
	"sort"
)

//create a Single Singature contract for owner  。
func CreateSignatureContract(ownerPubKey *crypto.PubKey) (*Contract,error){

	return &Contract{
		Code: CreateSignatureRedeemScript(ownerPubKey),
		Parameters: []ContractParameterType{Signature},
		OwnerPubkeyHash: ToCodeHash(ownerPubKey.EncodePoint(true)),
	},nil
}

func CreateSignatureRedeemScript(pubkey *crypto.PubKey) []byte{
	sb := pg.NewProgramBuilder()
	sb.PushData(pubkey.EncodePoint(true))
	sb.AddOp(vm.OP_CHECKSIG)
	return sb.ToArray()
}

//create a Multi Singature contract for owner  。
func CreateMultiSigContract(publicKeyHash Uint160,m int, publicKeys []*crypto.PubKey) (*Contract,error){

	params := make([]ContractParameterType,m)
	for i,_ := range params{
		params[i] = Signature
	}

	return &Contract{
		Code: CreateMultiSigRedeemScript(m,publicKeys),
		Parameters: params,
		OwnerPubkeyHash: publicKeyHash,
	},nil
}

func CreateMultiSigRedeemScript(m int,pubkeys []*crypto.PubKey) []byte{
	if ! (m >= 1 && m <= len(pubkeys) && len(pubkeys) <= 24) {
		return nil //TODO: add panic
	}

	sb := pg.NewProgramBuilder()
	sb.PushNumber(big.NewInt(int64(m)))

	//sort pubkey
	sort.Sort(crypto.PubKeySlice(pubkeys))

	for _,pubkey := range pubkeys{
		sb.PushData(pubkey.EncodePoint(true))
	}

	sb.PushNumber(big.NewInt(int64(len(pubkeys))))
	sb.AddOp(vm.OP_CHECKMULTISIG)
	return sb.ToArray()
}
