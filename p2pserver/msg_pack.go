package p2pserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	core "github.com/Ontology/core/types"
	actor "github.com/Ontology/p2pserver/actor/req"
	msgCommon "github.com/Ontology/p2pserver/common"
	types "github.com/Ontology/p2pserver/common"
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

func NewVersion(n *peer.Peer) ([]byte, error) {
	log.Debug()
	var msg msg.Version
	vpl := constructVersionPayload(n)
	msg.P = vpl
	msg.PK = n.GetPubKey()
	log.Debug("new version msg.pk is ", msg.PK)

	msg.Hdr.Magic = msgCommon.NETMAGIC
	copy(msg.Hdr.CMD[0:7], "version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	msg.PK.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.Hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewPingMsg(height uint64) ([]byte, error) {
	var msg msg.Ping
	msg.Hdr.Magic = types.NETMAGIC
	copy(msg.Hdr.CMD[0:7], "ping")
	msg.Height = height
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.Height)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewHeadersReq() ([]byte, error) {
	var h msg.HeadersReq
	h.P.Len = 1
	buf, _ := actor.GetCurrentHeaderHash()
	copy(h.P.HashEnd[:], buf[:])

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new HeadersReq")
		return nil, err
	}

	s := msg.CheckSum(p.Bytes())
	h.Hdr.Init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
	return []byte{}, nil
}

func NewHeaders(headers []core.Header, count uint32) ([]byte, error) {
	var msg msg.BlkHeader
	msg.Cnt = count
	msg.BlkHdr = headers
	msg.Hdr.Magic = types.NETMAGIC
	cmd := "headers"
	copy(msg.Hdr.CMD[0:len(cmd)], cmd)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, msg.Cnt)
	for _, header := range headers {
		header.Serialize(tmpBuffer)
	}
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
