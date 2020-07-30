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
	Timestamp   uint32
	View        uint32
	NodePubKeys []string
	Proposor    string

	ProposerSig []byte

	Voters []VoterMsg
}

type VoterMsg struct {
	offlineIndex []uint8
	PubKey       string
	Sig          []byte
}

func (this *OfflineWitnessMsg) CmdType() string {
	return common2.SUBNET_OFFLINE_TYPE
}

func (self *OfflineWitnessMsg) Serialization(sink *common.ZeroCopySink) {
	self.serializeUnsigned(sink)
	sink.WriteVarBytes(self.ProposerSig)

	sink.WriteUint32(uint32(len(self.Voters)))
	for _, val := range self.Voters {
		sink.WriteVarBytes(val.offlineIndex)
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

	self.Proposor, err = source.ReadString()
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

		self.Voters = append(self.Voters, struct {
			offlineIndex []uint8
			PubKey       string
			Sig          []byte
		}{offlineIndex: index, PubKey: pubKey, Sig: sig})
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
	sink.WriteString(self.Proposor)
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
	self.Voters = append(self.Voters, struct {
		offlineIndex []uint8
		PubKey       string
		Sig          []byte
	}{offlineIndex: index, PubKey: pubkey, Sig: sig})

	return nil
}

func (self *OfflineWitnessMsg) VerifySigs() error {
	sink := common.NewZeroCopySink(nil)
	self.serializeUnsigned(sink)
	unsign := sink.Bytes()
	data := sha256.Sum256(unsign)
	prop, err := vconfig.Pubkey(self.Proposor)
	if err != nil {
		return err
	}

	err = signature.Verify(prop, data[:], self.ProposerSig)
	if err != nil {
		return err
	}

	for _, vote := range self.Voters {
		sink = common.NewZeroCopySink(unsign)
		sink.WriteVarBytes(vote.offlineIndex)
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
