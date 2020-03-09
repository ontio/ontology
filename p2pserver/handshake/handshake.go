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
package handshake

import (
	"fmt"
	"net"
	"time"

	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
)

const HANDSHAKE_DURATION = 10 * time.Second // handshake time can not exceed this duration, or will treat as attack.

func HandshakeClient(info *peer.PeerInfo, selfId *kbucket.KadKeyId, conn net.Conn) (*peer.PeerInfo, error) {
	version := newVersion(info)
	if err := conn.SetDeadline(time.Now().Add(HANDSHAKE_DURATION)); err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{}) //reset back
	}()

	// 1. sendMsg version
	err := sendMsg(conn, version)
	if err != nil {
		return nil, err
	}

	// 2. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		return nil, err
	}
	receivedVersion, ok := msg.(*types.Version)
	if !ok {
		return nil, fmt.Errorf("expected version message, but got message type: %s", msg.CmdType())
	}

	// 3. update kadId
	kid := kbucket.PseudoKadIdFromUint64(receivedVersion.P.Nonce)
	if receivedVersion.P.SoftVersion > "v1.9.0" {
		err = sendMsg(conn, &types.UpdateKadId{KadKeyId: selfId})
		if err != nil {
			return nil, err
		}
		// 4. read kadkeyid
		msg, _, err = types.ReadMessage(conn)
		if err != nil {
			return nil, err
		}
		kadKeyId, ok := msg.(*types.UpdateKadId)
		if !ok {
			return nil, fmt.Errorf("handshake failed, expect kad id message, got %s", msg.CmdType())
		}

		kid = kadKeyId.KadKeyId.Id
	}

	// 5. sendMsg ack
	err = sendMsg(conn, &types.VerACK{})
	if err != nil {
		return nil, err
	}

	return createPeerInfo(receivedVersion, kid), nil
}

func HandshakeServer(info *peer.PeerInfo, selfId *kbucket.KadKeyId, conn net.Conn) (*peer.PeerInfo, error) {
	ver := newVersion(info)
	if err := conn.SetDeadline(time.Now().Add(HANDSHAKE_DURATION)); err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{}) //reset back
	}()

	// 1. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERSION_TYPE {
		return nil, fmt.Errorf("[HandshakeServer] expected version message")
	}
	version := msg.(*types.Version)

	// 2. sendMsg version
	err = sendMsg(conn, ver)
	if err != nil {
		return nil, err
	}

	// 3. read update kadkey id
	kid := kbucket.PseudoKadIdFromUint64(version.P.Nonce)
	if version.P.SoftVersion > "v1.9.0" {
		msg, _, err := types.ReadMessage(conn)
		if err != nil {
			return nil, fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		}
		kadkeyId, ok := msg.(*types.UpdateKadId)
		if !ok {
			return nil, fmt.Errorf("[HandshakeServer] expected update kadkeyid message")
		}
		kid = kadkeyId.KadKeyId.Id
		// 4. sendMsg update kadkey id
		err = sendMsg(conn, &types.UpdateKadId{KadKeyId: selfId})
		if err != nil {
			return nil, err
		}
	}

	// 5. read version ack
	msg, _, err = types.ReadMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERACK_TYPE {
		return nil, fmt.Errorf("[HandshakeServer] expected version ack message")
	}

	return createPeerInfo(version, kid), nil
}

func sendMsg(conn net.Conn, msg types.Message) error {
	sink := common2.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)
	_, err := conn.Write(sink.Bytes())
	if err != nil {
		return fmt.Errorf("[handshake]error sending messge to %s :%s", conn.LocalAddr(), err.Error())
	}

	return nil
}

func createPeerInfo(version *types.Version, kid kbucket.KadId) *peer.PeerInfo {
	return peer.NewPeerInfo(kid, version.P.Version, version.P.Services, version.P.Relay != 0, version.P.HttpInfoPort,
		version.P.SyncPort, version.P.StartHeight, version.P.SoftVersion)
}

func newVersion(peerInfo *peer.PeerInfo) *types.Version {
	var version types.Version
	version.P = types.VersionPayload{
		Version:      peerInfo.Version,
		Services:     peerInfo.Services,
		SyncPort:     peerInfo.Port,
		Nonce:        peerInfo.Id.ToUint64(),
		IsConsensus:  false,
		HttpInfoPort: peerInfo.HttpInfoPort,
		StartHeight:  peerInfo.Height,
		TimeStamp:    time.Now().UnixNano(),
		SoftVersion:  peerInfo.SoftVersion,
	}

	if peerInfo.Relay {
		version.P.Relay = 1
	} else {
		version.P.Relay = 0
	}
	if peerInfo.HttpInfoPort > 0 {
		version.P.Cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		version.P.Cap[common.HTTP_INFO_FLAG] = 0x00
	}

	return &version
}
