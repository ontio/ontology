package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/log"
	types "github.com/Ontology/p2pserver/common"
)

type NodeAddr struct {
	Time          int64
	Services      uint64
	IpAddr        [16]byte
	Port          uint16
	ConsensusPort uint16
	ID            uint64 // Unique ID
}

type Addr struct {
	Hdr       MsgHdr
	NodeCnt   uint64
	NodeAddrs []types.PeerAddr
}


func (msg Addr) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Addr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg.Hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, msg.NodeCnt)
	if err != nil {
		return nil, err
	}
	for _, v := range msg.NodeAddrs {
		err = binary.Write(&buf, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), err
}

func (msg *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.NodeCnt))
	log.Debug("The address count is ", msg.NodeCnt)
	msg.NodeAddrs = make([]types.PeerAddr, msg.NodeCnt)
	for i := 0; i < int(msg.NodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.NodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}
