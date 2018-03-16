package types

import (
	"bytes"
	"crypto/sha256"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/common"
	"golang.org/x/crypto/ripemd160"
	"errors"
	"github.com/Ontology/common/serialization"
)

func AddressFromPubKey(pubkey *crypto.PubKey) Uint160 {
	buf := bytes.Buffer{}
	pubkey.Serialize(&buf)

	var u160 Uint160
	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(u160[:0])

	u160[0] = 0x01

	return u160
}

func AddressFromMultiPubKeys(pubkeys []*crypto.PubKey, m int) (Uint160, error) {
	var u160 Uint160
	n := len(pubkeys)
	if m <= 0 || m > n || n > 24 {
		return u160, errors.New("wrong multi-sig param")
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
	md.Sum(u160[:0])
	u160[0] = 0x02

	return u160, nil
}

func AddressFromBookKeepers(bookKeepers []*crypto.PubKey) (Uint160, error) {
	return AddressFromMultiPubKeys(bookKeepers, len(bookKeepers)-(len(bookKeepers)-1)/3)
}

