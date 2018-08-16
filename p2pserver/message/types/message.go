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

package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
)

type Message interface {
	Serialization(sink *comm.ZeroCopySink) error
	Deserialization(source *comm.ZeroCopySource) error
	CmdType() string
}

//MsgPayload in link channel
type MsgPayload struct {
	Id          uint64  //peer ID
	Addr        string  //link address
	PayloadSize uint32  //payload size
	Payload     Message //msg payload
}

type messageHeader struct {
	Magic    uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

func readMessageHeader(reader io.Reader) (messageHeader, error) {
	msgh := messageHeader{}
	err := binary.Read(reader, binary.LittleEndian, &msgh)
	return msgh, err
}

func writeMessageHeaderInto(sink *comm.ZeroCopySink, msgh messageHeader) {
	sink.WriteUint32(msgh.Magic)
	sink.WriteBytes(msgh.CMD[:])
	sink.WriteUint32(msgh.Length)
	sink.WriteBytes(msgh.Checksum[:])
}

func writeMessageHeader(writer io.Writer, msgh messageHeader) error {
	return binary.Write(writer, binary.LittleEndian, msgh)
}

func newMessageHeader(cmd string, length uint32, checksum [common.CHECKSUM_LEN]byte) messageHeader {
	msgh := messageHeader{}
	msgh.Magic = config.DefConfig.P2PNode.NetworkMagic
	copy(msgh.CMD[:], cmd)
	msgh.Checksum = checksum
	msgh.Length = length
	return msgh
}

func WriteMessage(sink *comm.ZeroCopySink, msg Message) error {
	pstart := sink.Size()
	sink.NextBytes(common.MSG_HDR_LEN) // can not save the buf, since it may reallocate in sink
	err := msg.Serialization(sink)
	if err != nil {
		return err
	}
	pend := sink.Size()
	total := pend - pstart
	payLen := total - common.MSG_HDR_LEN

	sink.BackUp(total)
	buf := sink.NextBytes(total)
	checksum := common.Checksum(buf[common.MSG_HDR_LEN:])
	hdr := newMessageHeader(msg.CmdType(), uint32(payLen), checksum)

	sink.BackUp(total)
	writeMessageHeaderInto(sink, hdr)
	sink.NextBytes(payLen)

	return err
}

func ReadMessage(reader io.Reader) (Message, uint32, error) {
	hdr, err := readMessageHeader(reader)
	if err != nil {
		return nil, 0, err
	}

	magic := config.DefConfig.P2PNode.NetworkMagic
	if hdr.Magic != magic {
		return nil, 0, fmt.Errorf("unmatched magic number %d, expected %d", hdr.Magic, magic)
	}

	if hdr.Length > common.MAX_PAYLOAD_LEN {
		return nil, 0, fmt.Errorf("msg payload length:%d exceed max payload size: %d",
			hdr.Length, common.MAX_PAYLOAD_LEN)
	}

	buf := make([]byte, hdr.Length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, 0, err
	}

	checksum := common.Checksum(buf)
	if checksum != hdr.Checksum {
		return nil, 0, fmt.Errorf("message checksum mismatch: %x != %x ", hdr.Checksum, checksum)
	}

	cmdType := string(bytes.TrimRight(hdr.CMD[:], string(0)))
	msg, err := MakeEmptyMessage(cmdType)
	if err != nil {
		return nil, 0, err
	}

	// the buf is referenced by msg to avoid reallocation, so can not reused
	source := comm.NewZeroCopySource(buf)
	err = msg.Deserialization(source)
	if err != nil {
		return nil, 0, err
	}

	return msg, hdr.Length, nil
}

func MakeEmptyMessage(cmdType string) (Message, error) {
	switch cmdType {
	case common.PING_TYPE:
		return &Ping{}, nil
	case common.VERSION_TYPE:
		return &Version{}, nil
	case common.VERACK_TYPE:
		return &VerACK{}, nil
	case common.PONG_TYPE:
		return &Pong{}, nil
	case common.GET_HEADERS_TYPE:
		return &HeadersReq{}, nil
	case common.HEADERS_TYPE:
		return &BlkHeader{}, nil
	case common.INV_TYPE:
		return &Inv{}, nil
	case common.GET_DATA_TYPE:
		return &DataReq{}, nil
	case common.BLOCK_TYPE:
		return &Block{}, nil
	case common.TX_TYPE:
		return &Trn{}, nil
	case common.CONSENSUS_TYPE:
		return &Consensus{}, nil
	case common.NOT_FOUND_TYPE:
		return &NotFound{}, nil
	case common.DISCONNECT_TYPE:
		return &Disconnected{}, nil
	case common.GET_BLOCKS_TYPE:
		return &BlocksReq{}, nil
	case common.DHT_PING:
		return &DHTPing{}, nil
	case common.DHT_PONG:
		return &DHTPong{}, nil
	case common.DHT_FIND_NODE:
		return &FindNode{}, nil
	case common.DHT_NEIGHBORS:
		return &Neighbors{}, nil
	default:
		return nil, errors.New("unsupported cmd type:" + cmdType)
	}

}
