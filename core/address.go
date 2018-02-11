package core

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/Ontology/common"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"golang.org/x/crypto/ripemd160"
)


func AddressFromPubKey(pubkey *crypto.PubKey) types.Address {
	buf := bytes.Buffer{}
	pubkey.Serialize(&buf)

	var addr types.Address
	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])

	addr[0] = 0x01

	return addr
}

func AddressFromVmCode(vmCode types.VmCode) types.Address {
	var addr types.Address
	temp := sha256.Sum256(vmCode.Code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])

	addr[0] = byte(vmCode.CodeType)

	return addr
}

func AddressFromBookKeepers(bookKeepers []*crypto.PubKey) (types.Address, error) {
	if len(bookKeepers) < 1 {
		return types.Address{}, errors.New("[Ledger] , GetBookKeeperAddress with no bookKeeper")
	}
	var temp []byte
	var err error
	if len(bookKeepers) > 1 {
		temp, err = contract.CreateMultiSigRedeemScript(len(bookKeepers)-(len(bookKeepers)-1)/3, bookKeepers)
		if err != nil {
			return types.Address{}, errors.New("[Ledger],GetBookKeeperAddress failed with CreateMultiSigRedeemScript.")
		}
	} else {
		temp, err = contract.CreateSignatureRedeemScript(bookKeepers[0])
		if err != nil {
			return types.Address{}, errors.New("[Ledger],GetBookKeeperAddress failed with CreateMultiSigRedeemScript.")
		}
	}

	// TODO
	codehash := common.ToCodeHash(temp)
	return types.Address(codehash), nil
}
