package types

import (
	"bytes"
	"crypto/sha256"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/common"
	"golang.org/x/crypto/ripemd160"
	"errors"
	"github.com/Ontology/common/serialization"
	"sort"
)

func AddressFromPubKey(pubkey *crypto.PubKey) Address {
	buf := bytes.Buffer{}
	pubkey.Serialize(&buf)

	var addr Address
	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])

	addr[0] = 0x01

	return addr
}

func AddressFromMultiPubKeys(pubkeys []*crypto.PubKey, m int) (Address, error) {
	var addr Address
	n := len(pubkeys)
	if m <= 0 || m > n || n > 24 {
		return addr, errors.New("wrong multi-sig param")
	}
	sort.Sort(crypto.PubKeySlice(pubkeys))
	buf := bytes.Buffer{}
	serialization.WriteUint8(&buf, uint8(n))
	serialization.WriteUint8(&buf, uint8(m))
	for _, pubkey := range pubkeys {
		pubkey.Serialize(&buf)
	}

	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])
	addr[0] = 0x02

	return addr, nil
}

func AddressFromBookKeepers(bookKeepers []*crypto.PubKey) (Address, error) {
	return AddressFromMultiPubKeys(bookKeepers, len(bookKeepers)-(len(bookKeepers)-1)/3)
}

