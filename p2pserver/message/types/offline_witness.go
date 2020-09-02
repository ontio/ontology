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
	"crypto/sha256"
	"errors"
	"math"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	common2 "github.com/ontio/ontology/p2pserver/common"
)

type OfflineWitnessMsg struct {
	Timestamp   uint32   `json:"timestamp"`
	View        uint32   `json:"view"`
	NodePubKeys []string `json:"nodePubKeys"`
	Proposer    string   `json:"proposer"`

	ProposerSig []byte `json:"proposerSig"`

	Voters []VoterMsg `json:"voters"`
}

type VoterMsg struct {
	OfflineIndex []uint8 `json:"offlineIndex"`
	PubKey       string  `json:"pubKey"`
	Sig          []byte  `json:"sig"`
}

func (this *OfflineWitnessMsg) CmdType() string {
	return common2.SUBNET_OFFLINE_TYPE
}

func (self *OfflineWitnessMsg) Serialization(sink *common.ZeroCopySink) {
	self.serializeUnsigned(sink)
	sink.WriteVarBytes(self.ProposerSig)

	sink.WriteUint32(uint32(len(self.Voters)))
	for _, val := range self.Voters {
		sink.WriteVarBytes(val.OfflineIndex)
		sink.WriteString(val.PubKey)
		sink.WriteVarBytes(val.Sig)
	}
}

func (self *OfflineWitnessMsg) Deserialization(source *common.ZeroCopySource) (err error) {
	self.Timestamp, err = source.ReadUint32()
	if err != nil {
		return err
	}
	self.View, err = source.ReadUint32()
	if err != nil {
		return err
	}
	lenPubKeys, err := source.ReadUint32()
	if err != nil {
		return err
	}
	if lenPubKeys > math.MaxUint8 {
		return errors.New("too many node keys")
	}
	for i := uint32(0); i < lenPubKeys; i++ {
		key, err := source.ReadString()
		if err != nil {
			return err
		}
		self.NodePubKeys = append(self.NodePubKeys, key)
	}

	self.Proposer, err = source.ReadString()
	if err != nil {
		return err
	}

	lenVoters, err := source.ReadUint32()
	if err != nil {
		return err
	}

	for i := uint32(0); i < lenVoters; i++ {
		index, err := source.ReadVarBytes()
		if err != nil {
			return err
		}
		for _, idx := range index {
			if int(idx) >= len(self.NodePubKeys) {
				return errors.New("vote index out of range")
			}
		}
		pubKey, err := source.ReadString()
		if err != nil {
			return err
		}
		sig, err := source.ReadVarBytes()
		if err != nil {
			return err
		}

		self.Voters = append(self.Voters, VoterMsg{OfflineIndex: index, PubKey: pubKey, Sig: sig})
	}

	return self.VerifySigs()
}

func (self *OfflineWitnessMsg) serializeUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(self.Timestamp)
	sink.WriteUint32(self.View)
	sink.WriteUint32(uint32(len(self.NodePubKeys)))
	for _, key := range self.NodePubKeys {
		sink.WriteString(key)
	}
	sink.WriteString(self.Proposer)
}

func (self *OfflineWitnessMsg) Hash() common.Uint256 {
	sink := common.NewZeroCopySink(nil)
	self.serializeUnsigned(sink)
	hash := common.Uint256(sha256.Sum256(sink.Bytes()))

	return hash
}

func (self *OfflineWitnessMsg) AddProposeSig(acct *account.Account) error {
	hash := self.Hash()
	sig, err := signature.Sign(acct, hash[:])
	if err != nil {
		return err
	}
	self.ProposerSig = sig

	return nil
}

func (self *OfflineWitnessMsg) VoteFor(acct *account.Account, index []uint8) error {
	sink := common.NewZeroCopySink(nil)
	self.serializeUnsigned(sink)
	sink.WriteVarBytes(index)
	hash := common.Uint256(sha256.Sum256(sink.Bytes()))
	sig, err := signature.Sign(acct, hash[:])
	if err != nil {
		return err
	}
	pubkey := vconfig.PubkeyID(acct.PublicKey)
	self.Voters = append(self.Voters, VoterMsg{OfflineIndex: index, PubKey: pubkey, Sig: sig})

	return nil
}

func (self *OfflineWitnessMsg) VerifySigs() error {
	sink := common.NewZeroCopySink(nil)
	self.serializeUnsigned(sink)
	unsign := sink.Bytes()
	data := sha256.Sum256(unsign)
	prop, err := vconfig.Pubkey(self.Proposer)
	if err != nil {
		return err
	}

	err = signature.Verify(prop, data[:], self.ProposerSig)
	if err != nil {
		return err
	}

	for _, vote := range self.Voters {
		sink = common.NewZeroCopySink(unsign)
		sink.WriteVarBytes(vote.OfflineIndex)
		data = sha256.Sum256(sink.Bytes())
		key, err := vconfig.Pubkey(vote.PubKey)
		if err != nil {
			return err
		}
		err = signature.Verify(key, data[:], vote.Sig)
		if err != nil {
			return err
		}
	}

	return nil
}
