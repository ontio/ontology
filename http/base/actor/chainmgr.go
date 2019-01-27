/*
 * Copyright (C) 2019 The ontology Authors
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

package actor

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
)

var chainMgrPid *actor.PID

func SetChainMgrPid(actr *actor.PID) {
	chainMgrPid = actr
}

func AppendTxToRemoteShard(txn *types.Transaction) (errors.ErrCode, string) {

	ch := make(chan *shardmsg.TxResult, 1)
	txReq := &shardmsg.TxReq{txn, ch}
	chainMgrPid.Tell(txReq)
	if msg, ok := <-ch; ok {
		return msg.Err, msg.Desc
	}

	return errors.ErrUnknown, ""
}
