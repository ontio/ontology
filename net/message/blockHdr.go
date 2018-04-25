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

package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/net/actor"
	"github.com/ontio/ontology/net/protocol"
	"fmt"
)

type hdrHashReq struct {
	len       uint8
	hashStart [protocol.HASH_LEN]byte
	hashEnd   [protocol.HASH_LEN]byte
}

type headersReq struct {
	hdr msgHdr
	p   hdrHashReq
}

type blkHeader struct {
	hdr    msgHdr
	cnt    uint32
	blkHdr []types.Header
}

func NewHeadersReq() ([]byte, error) {
	var h headersReq

	h.p.len = 1
	buf, _ := actor.GetCurrentHeaderHash()
	copy(h.p.hashEnd[:], buf[:])
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Error("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.hdr.init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
}

func (msg headersReq) Verify(buf []byte) error {
	err := msg.hdr.Verify(buf)
	return err
}

func (msg blkHeader) Verify(buf []byte) error {
	err := msg.hdr.Verify(buf)
	return err
}

func (msg headersReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.p.len)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.p.hashStart)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, msg.p.hashEnd)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.len))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.hashStart))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.hashEnd))
	return err
}

func (msg blkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.cnt)
	if err != nil {
		return nil, err
	}

	for _, header := range msg.blkHdr {
		header.Serialize(buf)
	}
	return buf.Bytes(), err
}

func (msg *blkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.cnt))
	if err != nil {
		return err
	}

	for i := 0; i < int(msg.cnt); i++ {
		var headers types.Header
		err := (&headers).Deserialize(buf)
		msg.blkHdr = append(msg.blkHdr, headers)
		if err != nil {
			log.Debug("blkHeader Deserialization failed")
			goto blkHdrErr
		}
	}

blkHdrErr:
	return err
}

func (msg headersReq) Handle(node protocol.Noder) error {
	log.Debug()
	var startHash [protocol.HASH_LEN]byte
	var stopHash [protocol.HASH_LEN]byte
	startHash = msg.p.hashStart
	stopHash = msg.p.hashEnd
	headers, cnt, err := GetHeadersFromHash(startHash, stopHash)
	if err != nil || headers == nil {
		return err
	}
	buf, err := NewHeaders(headers, cnt)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}

func SendMsgSyncHeaders(node protocol.Noder) {
	buf, err := NewHeadersReq()
	if err != nil {
		log.Error("failed build a new headersReq")
	} else {
		go node.Tx(buf)
	}
}

func (msg blkHeader) Handle(node protocol.Noder) error {
	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < msg.cnt; i++ {
		blkHdr = append(blkHdr, &msg.blkHdr[i])
	}
	node.LocalNode().OnHeaderReceive(blkHdr)
	return nil
}

func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]types.Header, uint32, error) {
	var count uint32 = 0
	var empty [protocol.HASH_LEN]byte
	headers := []types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight, err := actor.GetCurrentHeaderHeight()
	if err != nil {
		return nil, 0, fmt.Errorf("GetCurrentHeaderHeight error:%s", err)
	}
	if startHash == empty {
		if stopHash == empty {
			if curHeight > protocol.MAX_BLK_HDR_CNT {
				count = protocol.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := actor.GetHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, 0, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if count > protocol.MAX_BLK_HDR_CNT {
				count = protocol.MAX_BLK_HDR_CNT
			}
		}
	} else {
		bkStart, err := actor.GetHeaderByHash(startHash)
		if err != nil || bkStart == nil {
			return nil, 0, err
		}
		startHeight = bkStart.Height
		if stopHash != empty {
			bkStop, err := actor.GetHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, 0, err
			}
			stopHeight = bkStop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, 0, errors.New("do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= protocol.MAX_BLK_HDR_CNT {
				count = protocol.MAX_BLK_HDR_CNT
				stopHeight = startHeight - protocol.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > protocol.MAX_BLK_HDR_CNT {
				count = protocol.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash, err := actor.GetBlockHashByHeight(stopHeight + i)
		if err != nil {
			log.Errorf("GetBlockHashByHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		if hash == common.UINT256_EMPTY{
			break
		}
		hd, err := actor.GetHeaderByHash(hash)
		if err != nil || hd == nil {
			log.Errorf("GetHeaderByHash failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		headers = append(headers, *hd)
	}

	return headers, count, nil
}

func NewHeaders(headers []types.Header, count uint32) ([]byte, error) {
	var msg blkHeader
	msg.cnt = count
	msg.blkHdr = headers
	msg.hdr.Magic = protocol.NET_MAGIC
	cmd := "headers"
	copy(msg.hdr.CMD[0:len(cmd)], cmd)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, msg.cnt)
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
	buf := bytes.NewBuffer(s[:protocol.CHECKSUM_LEN])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
