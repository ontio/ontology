package message

import (
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/ledger"
	. "DNA/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type pong struct {
	msgHdr
	height uint64
}

func NewPongMsg() ([]byte, error) {
	var msg pong
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "pong")
	msg.height = uint64(ledger.DefaultLedger.Store.GetHeaderHeight())
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.height)
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
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func (msg pong) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg pong) Handle(node Noder) error {
	node.SetHeight(msg.height)
	return nil
}

func (msg pong) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteUint64(buf, msg.height)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err

}

func (msg *pong) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.height, err = serialization.ReadUint64(buf)
	return err
}
