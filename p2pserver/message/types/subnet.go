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

package types

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/p2pserver/common"
)

type SubnetMembersRequest struct {
	From      common.PeerId
	To        common.PeerId
	Timestamp uint32
	PubKey    keypair.PublicKey // only valid if Timestamp != 0
	Sig       []byte
}

type SubnetMembers struct {
	Members []MemberInfo
}

func (self *SubnetMembers) String() string {
	val, _ := json.Marshal(self.Members)

	return string(val)
}

type MemberInfo struct {
	PubKey string // hex encoded
	Addr   string
}

func NewMembersRequestFromSeed() *SubnetMembersRequest {
	return &SubnetMembersRequest{}
}

func (self *SubnetMembersRequest) FromSeed() bool {
	return self.Timestamp == 0
}

func (self *SubnetMembersRequest) Role() string {
	if self.FromSeed() {
		return "seed"
	}

	return "gov"
}

func NewMembersRequest(from, to common.PeerId, acc *account.Account) (*SubnetMembersRequest, error) {
	request := &SubnetMembersRequest{
		From:      from,
		To:        to,
		Timestamp: uint32(time.Now().Unix()),
		PubKey:    acc.PublicKey,
	}

	sig, err := signature.Sign(acc, request.sigdata())
	if err != nil {
		return nil, err
	}
	request.Sig = sig
	return request, nil
}

func (self *SubnetMembersRequest) sigdata() []byte {
	sink := comm.NewZeroCopySink(nil)
	self.From.Serialization(sink)
	self.To.Serialization(sink)
	sink.WriteUint32(self.Timestamp)
	return sink.Bytes()
}

//Serialize message payload
func (self *SubnetMembersRequest) Serialization(sink *comm.ZeroCopySink) {
	self.From.Serialization(sink)
	self.To.Serialization(sink)
	sink.WriteUint32(self.Timestamp)
	if self.Timestamp != 0 {
		sink.WriteVarBytes(keypair.SerializePublicKey(self.PubKey))
		sink.WriteVarBytes(self.Sig)
	}
}

func (self *SubnetMembersRequest) CmdType() string {
	return common.GET_SUBNET_MEMBERS_TYPE
}

//Deserialize message payload
func (self *SubnetMembersRequest) Deserialization(source *comm.ZeroCopySource) (err error) {
	err = self.From.Deserialization(source)
	if err != nil {
		return err
	}
	err = self.To.Deserialization(source)
	if err != nil {
		return err
	}
	self.Timestamp, err = source.ReadUint32()
	if err != nil {
		return err
	}
	if self.Timestamp != 0 {
		pubKey, e := source.ReadVarBytes()
		if e != nil {
			err = e
			return err
		}
		self.PubKey, err = keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return err
		}
		self.Sig, err = source.ReadVarBytes()
		if err != nil {
			return err
		}

		dur := time.Hour
		if uint32(time.Now().Add(-dur).Unix()) > self.Timestamp {
			return errors.New("subnet members request message expired")
		}
		err = signature.Verify(self.PubKey, self.sigdata(), self.Sig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *SubnetMembers) Serialization(sink *comm.ZeroCopySink) {
	sink.WriteUint32(uint32(len(self.Members)))
	for _, member := range self.Members {
		sink.WriteString(member.PubKey)
		sink.WriteString(member.Addr)
	}
}

func (self *SubnetMembers) CmdType() string {
	return common.SUBNET_MEMBERS_TYPE
}

//Deserialize message payload
func (self *SubnetMembers) Deserialization(source *comm.ZeroCopySource) error {
	num, err := source.ReadUint32()
	if err != nil {
		return err
	}
	var members []MemberInfo
	for i := uint32(0); i < num; i++ {
		pubKey, err := source.ReadString()
		if err != nil {
			return err
		}
		addr, err := source.ReadString()
		if err != nil {
			return err
		}

		members = append(members, MemberInfo{PubKey: pubKey, Addr: addr})
	}

	self.Members = members
	return nil
}
