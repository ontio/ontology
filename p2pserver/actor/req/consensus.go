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

package req

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/config"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
)

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}

func NotifyPeerState(peer keypair.PublicKey, connected bool) {
	if config.DefConfig.Genesis.ConsensusType != config.CONSENSUS_TYPE_VBFT {
		return
	}

	if ConsensusPid != nil {
		ConsensusPid.Tell(&msgTypes.PeerStateUpdate{
			PeerPubKey: peer,
			Connected:  connected,
		})
	}
}
