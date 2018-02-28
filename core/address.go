package core

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	vmtypes "github.com/Ontology/vm/types"
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

func AddressFromMultiPubKeys(pubkeys []*crypto.PubKey, m int) (types.Address, error) {
	var addr types.Address
	n := len(pubkeys)
	if m <= 0 || m > n || n > 24 {
		return addr, errors.New("wrong multi-sig param")
	}
	buf := bytes.Buffer{}
	serialization.WriteUint8(&buf, uint8(n))
	serialization.WriteUint8(&buf, uint8(m))
	for _, pubkey := range pubkeys {
		pubkey.Serialize(&buf)
	}

	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])
	addr[0] = 0x02

	return addr, nil
}

func AddressFromVmCode(vmCode vmtypes.VmCode) types.Address {
	var addr types.Address
	temp := sha256.Sum256(vmCode.Code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])

	addr[0] = byte(vmCode.CodeType)

	return addr
}

func AddressFromBookKeepers(bookKeepers []*crypto.PubKey) (types.Address, error) {
	return AddressFromMultiPubKeys(bookKeepers, len(bookKeepers)-(len(bookKeepers)-1)/3)
}
