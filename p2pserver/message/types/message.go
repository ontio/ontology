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
	"fmt"
	"io"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
)

type Message interface {
	Serialization(sink *comm.ZeroCopySink)
	Deserialization(source *comm.ZeroCopySource) error
	CmdType() string
}

type UnknownMessage struct {
	Cmd     string
	Payload []byte
}

func (self *UnknownMessage) CmdType() string {
	return self.Cmd
}

func (self *UnknownMessage) Serialization(sink *comm.ZeroCopySink) {
	sink.WriteBytes(self.Payload)
}

func (self *UnknownMessage) Deserialization(source *comm.ZeroCopySource) error {
	self.Payload, _ = source.NextBytes(source.Len())
	return nil
}

//MsgPayload in link channel
type MsgPayload struct {
	Id          common.PeerId //peer ID
	Addr        string        //link address
	PayloadSize uint32        //payload size
	Payload     Message       //msg payload
}

type messageHeader struct {
	Magic    uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

func readMessageHeader(reader io.Reader) (messageHeader, error) {
	msgh := messageHeader{}
	hdrBytes := make([]byte, comm.UINT32_SIZE+common.MSG_CMD_LEN+comm.UINT32_SIZE+common.CHECKSUM_LEN)

	if _, err := io.ReadFull(reader, hdrBytes); err != nil {
		return msgh, err
	}

	source := comm.NewZeroCopySource(hdrBytes)
	msgh.Magic, _ = source.NextUint32()
	cmd, _ := source.NextBytes(common.MSG_CMD_LEN)
	copy(msgh.CMD[:], cmd)
	msgh.Length, _ = source.NextUint32()
	checksum, _ := source.NextBytes(common.CHECKSUM_LEN)
	copy(msgh.Checksum[:], checksum)
	return msgh, nil
}

func writeMessageHeaderInto(sink *comm.ZeroCopySink, msgh messageHeader) {
	sink.WriteUint32(msgh.Magic)
	sink.WriteBytes(msgh.CMD[:])
	sink.WriteUint32(msgh.Length)
	sink.WriteBytes(msgh.Checksum[:])
}

func newMessageHeader(cmd string, length uint32, checksum [common.CHECKSUM_LEN]byte) messageHeader {
	msgh := messageHeader{}
	msgh.Magic = config.DefConfig.P2PNode.NetworkMagic
	copy(msgh.CMD[:], cmd)
	msgh.Checksum = checksum
	msgh.Length = length
	return msgh
}

func WriteMessage(sink *comm.ZeroCopySink, msg Message) {
	pstart := sink.Size()
	sink.NextBytes(common.MSG_HDR_LEN) // can not save the buf, since it may reallocate in sink
	msg.Serialization(sink)
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
	msg := makeEmptyMessage(cmdType)

	// the buf is referenced by msg to avoid reallocation, so can not reused
	source := comm.NewZeroCopySource(buf)
	err = msg.Deserialization(source)
	if err != nil {
		return nil, 0, err
	}

	return msg, hdr.Length, nil
}

func makeEmptyMessage(cmdType string) Message {
	switch cmdType {
	case common.PING_TYPE:
		return &Ping{}
	case common.VERSION_TYPE:
		return &Version{}
	case common.VERACK_TYPE:
		return &VerACK{}
	case common.ADDR_TYPE:
		return &Addr{}
	case common.GetADDR_TYPE:
		return &AddrReq{}
	case common.PONG_TYPE:
		return &Pong{}
	case common.GET_HEADERS_TYPE:
		return &HeadersReq{}
	case common.HEADERS_TYPE:
		return &BlkHeader{}
	case common.INV_TYPE:
		return &Inv{}
	case common.GET_DATA_TYPE:
		return &DataReq{}
	case common.BLOCK_TYPE:
		return &Block{}
	case common.TX_TYPE:
		return &Trn{}
	case common.CONSENSUS_TYPE:
		return &Consensus{}
	case common.NOT_FOUND_TYPE:
		return &NotFound{}
	case common.GET_BLOCKS_TYPE:
		return &BlocksReq{}
	case common.FINDNODE_TYPE:
		return &FindNodeReq{}
	case common.FINDNODE_RESP_TYPE:
		return &FindNodeResp{}
	case common.UPDATE_KADID_TYPE:
		return &UpdatePeerKeyId{}
	case common.GET_SUBNET_MEMBERS_TYPE:
		return &SubnetMembersRequest{}
	case common.SUBNET_MEMBERS_TYPE:
		return &SubnetMembers{}
	default:
		return &UnknownMessage{Cmd: cmdType}
	}
}
