package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"github.com/Ontology/common/log"
	types "github.com/Ontology/p2pserver/common"
)

type AddrReq struct {
	Hdr msgHdr
}
type NodeAddr struct {
	Time          int64
	Services      uint64
	IpAddr        [16]byte
	Port          uint16
	ConsensusPort uint16
	ID            uint64 // Unique ID
}

type Addr struct {
	hdr       msgHdr
	nodeCnt   uint64
	nodeAddrs []types.PeerAddr
}

const (
	NODEADDRSIZE = 30
)

func NewGetAddr() ([]byte, error) {
	var msg AddrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message get addr length is: ", len(buf), " ", str)
	return buf, err
}

func NewAddrs(nodeAddrs []types.PeerAddr, count uint64) ([]byte, error) {
	var msg Addr
	msg.nodeAddrs = nodeAddrs
	msg.nodeCnt = count
	msg.hdr.Magic = types.NETMAGIC
	cmd := "addr"
	copy(msg.hdr.CMD[0:7], cmd)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, msg.nodeAddrs)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func (msg AddrReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg AddrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (msg *AddrReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg Addr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg.hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		return nil, err
	}
	for _, v := range msg.nodeAddrs {
		err = binary.Write(&buf, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), err
}

func (msg *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.nodeCnt))
	log.Debug("The address count is ", msg.nodeCnt)
	msg.nodeAddrs = make([]types.PeerAddr, msg.nodeCnt)
	for i := 0; i < int(msg.nodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.nodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}

func (msg *NodeAddr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg NodeAddr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg Addr) Verify(buf []byte) error {
	err := msg.hdr.Verify(buf)
	// TODO Verify the message Content, check the ipaddr number
	return err
}
