package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
	vmtypes "github.com/Ontology/vm/types"
	"golang.org/x/crypto/ripemd160"
)

const AddrLen = 20

type Address [AddrLen]byte


func (self *Address) ToHexString() string {
	return fmt.Sprintf("%x", self[:])
}

func (self *Address) Serialize(w io.Writer) error {
	_, err := w.Write(self[:])
	return err
}

func (self *Address) Deserialize(r io.Reader) error {
	n, err := r.Read(self[:])
	if n != len(self[:]) || err != nil {
		return errors.New("deserialize Address error")
	}
	return nil
}


func AddressFromPubKey(pubkey *crypto.PubKey) Address {
	buf := bytes.Buffer{}
	pubkey.Serialize(&buf)

	var addr Address
	temp := sha256.Sum256(buf.Bytes())
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])

	addr[0] = 0x01

	return addr
}

func AddressFromMultiPubKeys(pubkeys []*crypto.PubKey, m int) (Address, error) {
	var addr Address
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

func AddressFromVmCode(vmCode vmtypes.VmCode) Address {
	var addr Address
	temp := sha256.Sum256(vmCode.Code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:])

	addr[0] = byte(vmCode.CodeType)

	return addr
}

func AddressFromBookKeepers(bookKeepers []*crypto.PubKey) (Address, error) {
	return AddressFromMultiPubKeys(bookKeepers, len(bookKeepers)-(len(bookKeepers)-1)/3)
}
