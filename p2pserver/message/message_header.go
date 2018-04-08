package message

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Ontology/common/log"
	"github.com/Ontology/p2pserver/common"
)

// The network communication message header
type MsgHdr struct {
	Magic uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

func (hdr *MsgHdr) Init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = common.NETMAGIC
	copy(hdr.CMD[0:uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:common.CHECKSUM_LEN])
	hdr.Length = length
}

// Verify the message header information
// @p payload of the message
func (hdr MsgHdr) Verify(buf []byte) error {
	if magicVerify(hdr.Magic) == false {
		log.Warn(fmt.Sprintf("Unmatched magic number 0x%0x", hdr.Magic))
		return errors.New("Unmatched magic number ")
	}
	checkSum := CheckSum(buf)
	if bytes.Equal(hdr.Checksum[:], checkSum[:]) == false {
		str1 := hex.EncodeToString(hdr.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		log.Warn(fmt.Sprintf("Message Checksum error, Received checksum %s Wanted checksum: %s",
			str1, str2))
		return errors.New("Message Checksum error ")
	}

	return nil
}

// FIXME how to avoid duplicate serial/deserial function as
// most of them are the same
func (hdr MsgHdr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (hdr *MsgHdr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p[0:common.MSG_HDR_LEN])
	err := binary.Read(buf, binary.LittleEndian, hdr)
	return err
}

