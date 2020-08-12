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

package subnet

import (
	"errors"
	"math/rand"
	"time"

	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
)

const MinProposeDuration = uint32(1 * time.Hour / time.Second)
const ExpireOfflineMsgTime = uint32(2 * time.Hour / time.Second)
const DelayUpdateMsgTime = 5 * time.Second

func (self *SubNet) ProposeOffline(nodes []string) error {
	if self.acct == nil {
		return errors.New("only consensus node can propose offline witness")
	}
	key := vconfig.PubkeyID(self.acct.PublicKey)
	role, view := self.gov.GetNodeRoleAndView(key)
	if role != utils.ConsensusNode {
		return errors.New("only consensus node can propose offline witness")
	}
	now := uint32(time.Now().Unix())
	connected := make(map[string]bool)
	for _, val := range self.GetMembersInfo() {
		if val.Connected {
			connected[val.PubKey] = true
		}
	}
	var leftNodes []string
	for _, node := range nodes {
		if !self.gov.IsGovNode(node) {
			continue
		}
		if !connected[node] {
			leftNodes = append(leftNodes, node)
		}
	}

	if len(leftNodes) == 0 {
		log.Info("do not send offline witness proposal since all nodes are online")
		return nil
	}
	msg := &types.OfflineWitnessMsg{
		Timestamp:   now,
		View:        view,
		NodePubKeys: leftNodes,
		Proposer:    key,
	}
	err := msg.AddProposeSig(self.acct)
	if err != nil {
		return err
	}
	err = self.addProposol(msg)
	if err != nil {
		return err
	}

	self.unparker.Unpark()
	return nil
}

func (self *SubNet) sendOfflineWitness(net p2p.P2P) {
	var msgs []struct {
		status WitnessStatus
		rawMsg []byte
	}
	var peerIds []common.PeerId
	now := uint32(time.Now().Unix())
	self.lock.RLock()
	for hash, m := range self.offlineWitness {
		if m.Msg.Timestamp+ExpireOfflineMsgTime < now {
			delete(self.offlineWitness, hash)
		}
		if m.Status != UnchangedStatus {
			rawMsg := common2.SerializeToBytes(m.Msg)
			msgs = append(msgs, struct {
				status WitnessStatus
				rawMsg []byte
			}{status: UnchangedStatus, rawMsg: rawMsg})
		}
	}
	for _, p := range self.connected {
		peerIds = append(peerIds, p.Id)
	}
	self.lock.RUnlock()

	for _, msg := range msgs {
		switch msg.status {
		case NewStatus, UpdatedStatus:
			for _, peerId := range peerIds {
				if p := net.GetPeer(peerId); p != nil {
					_ = p.SendRaw(msg.rawMsg)
				}
			}
		}
	}
}

func (self *SubNet) Broadcast(net p2p.P2P, msg types.Message) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for _, p := range self.connected {
		net.SendTo(p.Id, msg)
	}
}

func (self *SubNet) OnOfflineWitnessMsg(ctx *p2p.Context, msg *types.OfflineWitnessMsg) {
	status := self.processOfflineWitnessMsg(ctx, msg)
	switch status {
	case NewStatus:
		self.unparker.Unpark()
	case UpdatedStatus:
		delay := int64(rand.Intn(1000)) * int64(DelayUpdateMsgTime) / 1000
		time.Sleep(time.Duration(delay))
		self.unparker.Unpark()
	}
}

func (self *SubNet) processOfflineWitnessMsg(ctx *p2p.Context, msg *types.OfflineWitnessMsg) WitnessStatus {
	now := uint32(time.Now().Unix())
	hash := msg.Hash()
	if msg.Timestamp+ExpireOfflineMsgTime < now {
		self.logger.Infof("receive expired witness msg: %s", hash.ToHexString())
		return UnchangedStatus
	}
	role, view := self.gov.GetNodeRoleAndView(msg.Proposer)
	if role != utils.ConsensusNode || view != msg.View {
		self.logger.Infof("receive expired witness msg: %s, {role: %d, view: %d}, current view: %d",
			hash.ToHexString(), role, msg.View, view)
		return UnchangedStatus
	}

	voters := make(map[string]types.VoterMsg)
	for _, voter := range msg.Voters {
		if !self.gov.IsGovNode(voter.PubKey) {
			self.logger.Infof("receive witness msg: %s with wrong voter: %s", hash.ToHexString(), voter.PubKey)
			return UnchangedStatus
		}
		voters[voter.PubKey] = voter
	}

	self.lock.Lock()
	defer self.lock.Unlock()
	offline := self.offlineWitness[msg.Hash()]
	if offline == nil {
		govNode := self.acct != nil && self.gov.IsGovNodePubKey(self.acct.PublicKey)
		if govNode {
			err := msg.VoteFor(self.acct, self.collectOfflineIndexLocked(msg.NodePubKeys))
			if err != nil {
				self.logger.Infof("vote for witness msg error: %s", err)
				return UnchangedStatus
			}
		}
		offline = &Offline{Status: NewStatus, Msg: msg}
		self.offlineWitness[msg.Hash()] = offline

		return NewStatus
	}
	for _, vote := range offline.Msg.Voters {
		delete(voters, vote.PubKey)
	}
	for _, voter := range voters {
		offline.Msg.Voters = append(offline.Msg.Voters, voter)
		offline.Status = UpdatedStatus
	}
	if len(voters) > 0 {
		return UpdatedStatus
	}

	return UnchangedStatus
}

func (self *SubNet) collectOfflineIndexLocked(nodes []string) []uint8 {
	connected := make(map[string]bool)
	for _, val := range self.cleanAndGetMembersInfoLocked() {
		if val.Connected {
			connected[val.PubKey] = true
		}
	}
	var leftNodes []uint8
	for idx, node := range nodes {
		if !connected[node] {
			leftNodes = append(leftNodes, uint8(idx))
		}
	}

	return leftNodes
}

func (self *SubNet) addProposol(msg *types.OfflineWitnessMsg) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, m := range self.offlineWitness {
		propose := m.Msg
		if propose.Proposer == msg.Proposer && (msg.Timestamp < propose.Timestamp+MinProposeDuration) {
			return errors.New("have already propose offline witness recently")
		}
	}

	self.offlineWitness[msg.Hash()] = &Offline{Status: NewStatus, Msg: msg}
	return nil
}

func (self *SubNet) GetOfflineVotes() []*types.OfflineWitnessMsg {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var voters []*types.OfflineWitnessMsg
	for _, m := range self.offlineWitness {
		msg := *m.Msg
		voters = append(voters, &msg)
	}

	return voters
}
