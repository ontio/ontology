package p2pserver

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actor "github.com/Ontology/p2pserver/actor/req"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
	"github.com/Ontology/p2pserver/peer"
)

func constructVersionPayload(p *peer.Peer) msg.VersionPayload {
	vpl := msg.VersionPayload{}
	vpl.Version = p.GetVersion()
	vpl.Services = p.GetServices()
	vpl.HttpInfoPort = p.GetHttpInfoPort()
	if config.Parameters.HttpInfoStart {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x01
	} else {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x00
	}

	vpl.TimeStamp = uint32(time.Now().UTC().UnixNano())
	vpl.Port = p.GetPort()
	vpl.Nonce = p.GetID()
	if p.GetRelay() {
		vpl.Relay = 1
	} else {
		vpl.Relay = 0
	}

	height, _ := actor.GetCurrentBlockHeight()
	vpl.StartHeight = uint64(height)

	return vpl
}

func VersionHandle(data msgCommon.MsgPayload, p2p *P2PServer) error {
	length := len(data.Payload)

	if length == 0 {
		log.Error(fmt.Sprintf("nil message for %s", msgCommon.VERSION_TYPE))
		return errors.New("nil message")
	}

	ver := msg.Version{}
	copy(ver.Hdr.CMD[0:len(msgCommon.VERSION_TYPE)], msgCommon.VERSION_TYPE)

	ver.Deserialization(data.Payload[:length])
	ver.Verify(data.Payload[msgCommon.MSG_HDR_LEN:length])

	localPeer := p2p.Self
	remotePeer := p2p.Self.Np.GetPeer(data.Id)

	if ver.P.Nonce == localPeer.GetID() {
		log.Warn("The node handshake with itself")
		return errors.New("The node handshake with itself")
	}

	s := remotePeer.GetState()
	if s != msgCommon.INIT && s != msgCommon.HAND {
		log.Warn("Unknow status to received version")
		return errors.New("Unknow status to received version")
	}

	// Obsolete node
	n, ret := localPeer.DelNbrNode(ver.P.Nonce)
	if ret == true {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", ver.P.Nonce))
		// Close the connection and release the node soure
		n.SetState(msgCommon.INACTIVITY)
		n.CloseConn()
	}

	log.Debug("handle version msg.pk is ", ver.PK)
	if ver.P.Cap[msg.HTTP_INFO_FLAG] == 0x01 {
		remotePeer.SetHttpInfoState(true)
	} else {
		remotePeer.SetHttpInfoState(false)
	}
	remotePeer.SetHttpInfoPort(ver.P.HttpInfoPort)
	remotePeer.SetBookkeeperAddr(ver.PK)
	// Todo: too much parameters, wrapper them as a paramter
	remotePeer.UpdateInfo(time.Now(), ver.P.Version, ver.P.Services,
		ver.P.Port, ver.P.Nonce, ver.P.Relay, ver.P.StartHeight)
	localPeer.AddNbrNode(remotePeer)

	var buf []byte
	if s == msgCommon.INIT {
		remotePeer.SetState(msgCommon.HANDSHAKE)
		versionPayload := constructVersionPayload(localPeer)
		buf, _ = msg.NewVersion(versionPayload, localPeer.GetPubKey())
	} else if s == msgCommon.HAND {
		remotePeer.SetState(msgCommon.HANDSHAKED)
		buf, _ = msg.NewVerack(false)
	}
	remotePeer.Send(buf, false)

	return nil
}
