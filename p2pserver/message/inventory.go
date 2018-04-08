package message

import (
	"io"
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	p2pCommon "github.com/Ontology/p2pserver/common"
)

var LastInvHash common.Uint256

type InvPayload struct {
	InvType common.InventoryType
	Cnt     uint32
	Blk     []byte
}

type Inv struct {
	Hdr MsgHdr
	P   InvPayload
}

func (msg *InvPayload) Serialization(w io.Writer) {
	serialization.WriteUint8(w, uint8(msg.InvType))
	serialization.WriteUint32(w, msg.Cnt)

	binary.Write(w, binary.LittleEndian, msg.Blk)
}

func (msg Inv) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Inv) invType() common.InventoryType {
	return msg.P.InvType
}

func (msg Inv) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.P.Serialization(buf)

	return buf.Bytes(), err
}

func (msg *Inv) Deserialization(p []byte) error {
	err := msg.Hdr.Deserialization(p)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(p[p2pCommon.MSG_HDR_LEN:])
	invType, err := serialization.ReadUint8(buf)
	if err != nil {
		return err
	}
	msg.P.InvType = common.InventoryType(invType)
	msg.P.Cnt, err = serialization.ReadUint32(buf)
	if err != nil {
		return err
	}

	msg.P.Blk = make([]byte, msg.P.Cnt*p2pCommon.HASH_LEN)
	err = binary.Read(buf, binary.LittleEndian, &(msg.P.Blk))

	return err
}