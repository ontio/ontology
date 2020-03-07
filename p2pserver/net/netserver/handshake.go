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
package netserver

import (
	"fmt"
	"net"
	"strings"
	"time"

	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
	"strconv"
)

func HandshakeClient(netServer *NetServer, conn net.Conn) error {
	addr := conn.RemoteAddr().String()
	log.Debugf("[p2p]peer %s connect with %s with %s",
		conn.LocalAddr().String(), conn.RemoteAddr().String(),
		conn.RemoteAddr().Network())

	// 1. sendMsg version
	version := msgpack.NewVersion(netServer, ledger.DefLedger.GetCurrentBlockHeight())
	err := sendMsg(conn, version)
	if err != nil {
		return err
	}

	// 2. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		return err
	}
	if msg.CmdType() != common.VERSION_TYPE {
		return fmt.Errorf("expected version message, but got message type: %s", msg.CmdType())
	}

	receivedVersion := msg.(*types.Version)
	remoteAddr := conn.RemoteAddr().String()
	if err = isHandWithSelf(netServer, remoteAddr, receivedVersion); err != nil {
		return err
	}

	// 3. update kadId
	if receivedVersion.P.SoftVersion > "v1.9.0" {
		log.Info("*******come in dht*******")
		msg := msgpack.NewUpdateKadKeyId(netServer)
		err = sendMsg(conn, msg)
		if err != nil {
			return err
		}
		// 4. read kadkeyid
		msg, _, err = types.ReadMessage(conn)
		if err != nil {
			return err
		}
		if msg.CmdType() != common.UPDATE_KADID_TYPE {
			return fmt.Errorf("")
		}
		kadKeyId := msg.(*types.UpdateKadId)
		if !netServer.UpdateDHT(kadKeyId.KadKeyId.Id) {
			log.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kadKeyId.KadKeyId.Id.ToHexString())
			return fmt.Errorf("[HandshakeClient] UpdateDHT failed, kadId: %s", kadKeyId.KadKeyId.Id.ToHexString())
		}
	}

	// 5. sendMsg ack
	ack := msgpack.NewVerAck()
	err = sendMsg(conn, ack)
	if err != nil {
		netServer.RemoveFromOutConnRecord(addr)
		log.Warn(err)
		return err
	}

	// Obsolete node
	err = removeOldPeer(netServer, receivedVersion.P.Nonce, conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	remotePeer, err := createPeer(netServer, receivedVersion, conn)
	if err != nil {
		return err
	}

	netServer.AddOutConnRecord(addr)
	netServer.AddPeerAddress(addr, remotePeer)
	remotePeer.Link.SetAddr(addr)
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(remotePeer.GetID())
	remotePeer.AttachChan(netServer.NetChan)
	netServer.AddNbrNode(remotePeer)
	log.Infof("remotePeer.GetId():%d,addr: %s, link id: %d", remotePeer.GetID(), addr, remotePeer.Link.GetID())
	go remotePeer.Link.Rx()
	remotePeer.SetState(common.ESTABLISH)

	if netServer.pid != nil {
		input := &common.AppendPeerID{
			ID: receivedVersion.P.Nonce,
		}
		netServer.pid.Tell(input)
	}

	return nil
}

func isHandWithSelf(netServer *NetServer, remoteAddr string, receivedVersion *types.Version) error {
	addrIp, err := common.ParseIPAddr(remoteAddr)
	if err != nil {
		log.Warn(err)
		return err
	}
	nodeAddr := addrIp + ":" + strconv.Itoa(int(receivedVersion.P.SyncPort))
	if receivedVersion.P.Nonce == netServer.GetID() {
		log.Warn("[createPeer]the node handshake with itself:", remoteAddr)
		netServer.SetOwnAddress(nodeAddr)
		return fmt.Errorf("[createPeer]the node handshake with itself: %s", remoteAddr)
	}
	return nil
}

func HandshakeServer(netServer *NetServer, conn net.Conn) error {
	// 1. read version
	msg, _, err := types.ReadMessage(conn)
	if err != nil {
		log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERSION_TYPE {
		return fmt.Errorf("[HandshakeServer] expected version message")
	}
	version := msg.(*types.Version)
	if err = isHandWithSelf(netServer, conn.RemoteAddr().String(), version); err != nil {
		return err
	}

	// 2. sendMsg version
	ver := msgpack.NewVersion(netServer, ledger.DefLedger.GetCurrentBlockHeight())
	err = sendMsg(conn, ver)
	if err != nil {
		log.Errorf("[HandshakeServer] WriteMessage sendMsg failed, error: %s", err)
		return err
	}

	// 3. read update kadkey id
	kid := kbucket.PseudoKadIdFromUint64(version.P.Nonce)
	if version.P.SoftVersion > "v1.9.0" {
		msg, _, err := types.ReadMessage(conn)
		if err != nil {
			log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
			return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		}
		if msg.CmdType() != common.UPDATE_KADID_TYPE {
			return fmt.Errorf("[HandshakeServer] expected update kadkeyid message")
		}
		kadkeyId := msg.(*types.UpdateKadId)
		kid = kadkeyId.KadKeyId.Id
		// 4. sendMsg update kadkey id
		msg = msgpack.NewUpdateKadKeyId(netServer)
		err = sendMsg(conn, msg)
		if err != nil {
			return err
		}
	}

	// 5. read version ack
	msg, _, err = types.ReadMessage(conn)
	if err != nil {
		log.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
		return fmt.Errorf("[HandshakeServer] ReadMessage failed, error: %s", err)
	}
	if msg.CmdType() != common.VERACK_TYPE {
		return fmt.Errorf("[HandshakeServer] expected version ack message")
	}

	// Obsolete node
	err = removeOldPeer(netServer, version.P.Nonce, conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	netServer.dht.Update(kid)
	remotePeer, err := createPeer(netServer, version, conn)
	if err != nil {
		return err
	}
	remotePeer.SetKId(kid)
	addr := conn.RemoteAddr().String()
	remotePeer.Link.SetAddr(addr)
	remotePeer.Link.SetConn(conn)
	remotePeer.Link.SetID(remotePeer.GetID())
	remotePeer.AttachChan(netServer.NetChan)

	netServer.AddNbrNode(remotePeer)
	netServer.AddInConnRecord(addr)
	netServer.AddPeerAddress(addr, remotePeer)

	go remotePeer.Link.Rx()
	if netServer.pid != nil {
		input := &common.AppendPeerID{
			ID: version.P.Nonce,
		}
		netServer.pid.Tell(input)
	}
	return nil
}

func sendMsg(conn net.Conn, msg types.Message) error {
	sink := common2.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)
	rawPacket := sink.Bytes()
	nByteCnt := len(rawPacket)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	_ = conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
	_, err := conn.Write(rawPacket)
	if err != nil {
		log.Infof("[handshake]error sending messge to %s :%s", conn.LocalAddr(), err.Error())
		return err
	}
	return nil
}

func checkReservedPeers(remoteAddr string) error {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		found := false
		for _, addr := range config.DefConfig.P2PNode.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(remoteAddr, addr) {
				log.Debug("[createPeer]peer in reserved list", remoteAddr)
				found = true
				break
			}
		}
		if !found {
			log.Debug("[createPeer]peer not in reserved list,close", remoteAddr)
			return fmt.Errorf("the remote addr: %s not in ReservedPeers", remoteAddr)
		}
	}

	return nil
}

func createPeer(p2p *NetServer, version *types.Version, conn net.Conn) (*peer.Peer, error) {
	log.Infof("remoteAddr: %s, localAddr: %s", conn.RemoteAddr().String(), conn.LocalAddr().String())

	remotePeer := peer.NewPeer()
	if version.P.Cap[common.HTTP_INFO_FLAG] == 0x01 {
		remotePeer.SetHttpInfoState(true)
	} else {
		remotePeer.SetHttpInfoState(false)
	}
	remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)

	remotePeer.UpdateInfo(time.Now(), version.P.Version,
		version.P.Services, version.P.SyncPort, version.P.Nonce,
		version.P.Relay, version.P.StartHeight, version.P.SoftVersion)

	return remotePeer, nil
}

func removeOldPeer(p2p *NetServer, pid uint64, remoteAddr string) error {
	p := p2p.GetPeer(pid)
	if p != nil {
		ipOld, err := common.ParseIPAddr(p.GetAddr())
		if err != nil {
			log.Warn("[createPeer]exist peer %d ip format is wrong %s", pid, p.GetAddr())
			return fmt.Errorf("[createPeer]exist peer %d ip format is wrong %s", pid, p.GetAddr())
		}
		ipNew, err := common.ParseIPAddr(remoteAddr)
		if err != nil {
			log.Warn("[createPeer]connecting peer %d ip format is wrong %s, close", pid, remoteAddr)
			return fmt.Errorf("[createPeer]connecting peer %d ip format is wrong %s, close", pid, remoteAddr)
		}
		if ipNew == ipOld {
			//same id and same ip
			n, delOK := p2p.DelNbrNode(pid)
			if delOK {
				log.Infof("[createPeer]peer reconnect %d", pid, remoteAddr)
				// Close the connection and release the node source
				n.Close()
				if p2p.pid != nil {
					input := &common.RemovePeerID{
						ID: pid,
					}
					p2p.pid.Tell(input)
				}
			}
		} else {
			err := fmt.Errorf("[createPeer]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
			log.Warn(err)
			return err
		}
	}

	return nil
}
