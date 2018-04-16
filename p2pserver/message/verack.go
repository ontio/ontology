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

package message

import (
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/protocol"
)

type verACK struct {
	msgHdr
}

func NewVerack() ([]byte, error) {
	var msg verACK
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("verack", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message tx verack length is ", len(buf), ", ", str)

	return buf, err
}

func (msg verACK) Handle(node protocol.Noder) error {
	log.Debug()

	s := node.GetState()
	if s != protocol.HAND_SHAKE && s != protocol.HAND_SHAKED {
		log.Warn("Unknow status to received verack")
		return errors.New("Unknow status to received verack")
	}

	node.SetState(protocol.ESTABLISH)

	if s == protocol.HAND_SHAKE {
		buf, _ := NewVerack()
		node.Tx(buf)
	}

	node.DumpInfo()
	node.ReqNeighborList()
	addr := node.GetAddr()
	port := node.GetPort()
	nodeAddr := addr + ":" + strconv.Itoa(int(port))
	node.LocalNode().RemoveAddrInConnectingList(nodeAddr)
	return nil
}
