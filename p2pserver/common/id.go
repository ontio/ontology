/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/big"
	"math/bits"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

var Difficulty = 18 //bit

type PeerId struct {
	val common.Address
}

type PeerIDAddressPair struct {
	ID      PeerId
	Address string
}

func (self PeerId) IsEmpty() bool {
	return self.val == common.ADDRESS_EMPTY
}

func (self PeerId) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(self.val)
}

func (self *PeerId) Deserialization(source *common.ZeroCopySource) error {
	val, eof := source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.val = val
	return nil
}

func (self *PeerId) ToHexString() string {
	return self.val.ToHexString()
}

type PeerKeyId struct {
	PublicKey keypair.PublicKey

	Id PeerId
}

func (self PeerId) GenRandPeerId(prefix uint) PeerId {
	var ret PeerId
	rand.Read(ret.val[:])
	if prefix == 0 {
		ret.val[0] &= 0x7f
		// make first bit different
		if (0x80 & self.val[0]) == 0 {
			ret.val[0] |= 0x80
		}
		return ret
	}

	num := prefix / 8
	left := prefix % 8
	if num > uint(len(self.val[:])) {
		num = uint(len(self.val[:]))
	}

	if num > 0 {
		copy(ret.val[:num], self.val[:num])
	}
	// make all prefix bits same
	mask := (uint8(0xff) >> (8 - left)) << (8 - left)
	ret.val[num] &= (^mask)
	ret.val[num] |= (self.val[num] & mask)

	// and prefix + 1 different
	// clear prefix + 1
	ret.val[num] &= ^(1 << (8 - left - 1))
	// if prefix + 1 bit is 1 then we already satisfied
	// if prefix + 1 bit is 0, then we should set this bit
	if (self.val[num]>>(8-left-1))&1 == 0 {
		ret.val[num] |= (1 << (8 - left - 1))
	}

	return ret
}

func (self PeerId) ToUint64() uint64 {
	if self.IsPseudoPeerId() {
		nonce := binary.LittleEndian.Uint64(self.val[:8])
		return nonce
	}
	kid := new(big.Int).SetBytes(self.val[:])
	uint64Max := new(big.Int).SetUint64(math.MaxUint64)
	res := kid.Mod(kid, uint64Max)
	return res.Uint64()
}

func (self PeerId) IsPseudoPeerId() bool {
	for i := 8; i < len(self.val); i++ {
		if self.val[i] != 0 {
			return false
		}
	}
	return true
}

func PseudoPeerIdFromUint64(data uint64) PeerId {
	id := common.ADDRESS_EMPTY
	binary.LittleEndian.PutUint64(id[:], data)
	return PeerId{
		val: id,
	}
}

func (this *PeerKeyId) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(keypair.SerializePublicKey(this.PublicKey))
}

func (this *PeerKeyId) Deserialization(source *common.ZeroCopySource) error {
	data, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	pub, err := keypair.DeserializePublicKey(data)
	if err != nil {
		return err
	}
	if !validatePublicKey(pub) {
		return errors.New("invalid kad public key")
	}
	this.PublicKey = pub
	this.Id = peerIdFromPubkey(pub)
	return nil
}

func peerIdFromPubkey(pubKey keypair.PublicKey) PeerId {
	return PeerId{val: types.AddressFromPubKey(pubKey)}
}

func RandPeerKeyId() *PeerKeyId {
	var acc *account.Account
	for {
		acc = account.NewAccount("")
		if validatePublicKey(acc.PublicKey) {
			break
		}
	}
	kid := peerIdFromPubkey(acc.PublicKey)
	return &PeerKeyId{
		PublicKey: acc.PublicKey,
		Id:        kid,
	}
}

func validatePublicKey(pubKey keypair.PublicKey) bool {
	pub := keypair.SerializePublicKey(pubKey)
	res := sha256.Sum256(pub)
	hash := sha256.Sum256(res[:])
	limit := Difficulty >> 3
	for i := 0; i < limit; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	diff := Difficulty % 8
	if diff != 0 {
		x := hash[limit] >> uint8(8-diff)
		return x == 0
	}
	return true
}

func (self PeerId) Distance(b PeerId) [20]byte {
	var c PeerId
	for i := 0; i < len(self.val); i++ {
		c.val[i] = self.val[i] ^ b.val[i]
	}

	return c.val
}

// Closer returns true if a is closer to self than b is
func (self PeerId) Closer(a, b PeerId) bool {
	adist := self.Distance(a)
	bdist := self.Distance(b)

	return bytes.Compare(adist[:], bdist[:]) < 0
}

// CommonPrefixLen(cpl) calculate two ID's xor prefix 0
func CommonPrefixLen(a, b PeerId) int {
	dis := a.Distance(b)
	return zeroPrefixLen(dis[:])
}

// ZeroPrefixLen returns the number of consecutive zeroes in a byte slice.
func zeroPrefixLen(id []byte) int {
	for i, b := range id {
		if b != 0 {
			return i*8 + bits.LeadingZeros8(uint8(b))
		}
	}

	return len(id) * 8
}
