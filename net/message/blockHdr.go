package message

import (
	"GoOnchain/common"
	"GoOnchain/common/log"
	"GoOnchain/common/serialization"
	"GoOnchain/core/ledger"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

type headersReq struct {
	hdr msgHdr
	p   struct {
		len       uint8
		hashStart [HASHLEN]byte
		hashEnd   [HASHLEN]byte
	}
}

type blkHeader struct {
	hdr    msgHdr
	cnt    uint32
	blkHdr []ledger.Blockdata
}

func NewHeadersReq(n Noder) ([]byte, error) {
	var h headersReq

	// Fixme correct with the exactly request length
	h.p.len = 1
	buf := n.GetLedger().Blockchain.CurrentBlockHash()
	copy(h.p.hashStart[:], reverse(buf[:]))

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
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg blkHeader) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg headersReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg blkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	serialization.WriteUint32(buf, msg.cnt)
	for _, header := range msg.blkHdr {
		header.Serialize(buf)
	}
	return buf.Bytes(), err
}

func (msg *blkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.cnt))
	log.Debug("The block header count is ", msg.cnt)
	msg.blkHdr = make([]ledger.Blockdata, msg.cnt)
	for i := 0; i < int(msg.cnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.blkHdr[i]))
		if err != nil {
			goto blkHdrErr
		}
	}
blkHdrErr:
	return err
}

func (msg headersReq) Handle(node Noder) error {
	common.Trace()
	// lock
	var starthash [HASHLEN]byte //[]common.Uint256
	var stophash [HASHLEN]byte  //common.Uint256
	starthash = msg.p.hashStart
	stophash = msg.p.hashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt := GetHeadersFromHash(starthash, stophash) //(starthash[0], stophash)
	buf, _ := NewHeaders(headers, cnt)
	go node.Tx(buf)
	return nil
}

func (msg blkHeader) Handle(node Noder) error {
	common.Trace()
	for i := 0; i < int(msg.cnt); i++ {
		var header ledger.Header
		header.Blockdata = &msg.blkHdr[i]
		err := ledger.DefaultLedger.Store.SaveHeader(&header)
		if err != nil {
			log.Warn("Add block Header error")
			return errors.New("Add block Header error\n")
		}
	}
	return nil
}
func GetHeadersFromHash(starthash common.Uint256, stophash common.Uint256) ([]ledger.Blockdata, uint32) {
	var count uint32 = 0
	var empty [HASHLEN]byte
	var headers []ledger.Blockdata
	bkstart, _ := ledger.DefaultLedger.GetBlockWithHash(starthash)
	startheight := bkstart.Blockdata.Height
	var stopheight uint32
	if stophash != empty {
		bkstop, _ := ledger.DefaultLedger.GetBlockWithHash(stophash)
		stopheight = bkstop.Blockdata.Height
		count = startheight - stopheight
		if count >= MAXBLKHDRCNT {
			count = MAXBLKHDRCNT
			stopheight = startheight - MAXBLKHDRCNT
		}
	} else {
		if startheight > MAXBLKHDRCNT {
			count = MAXBLKHDRCNT
		} else {
			count = startheight
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHeight
		bk, _ := ledger.DefaultLedger.GetBlockWithHeight(stopheight + i)
		headers = append(headers, *bk.Blockdata)
		i++
	}

	return headers, count
}

func NewHeaders(headers []ledger.Blockdata, count uint32) ([]byte, error) {
	var msg blkHeader
	msg.cnt = count
	msg.blkHdr = headers
	msg.hdr.Magic = NETMAGIC
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
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
