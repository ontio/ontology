package message

import (
	"DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/ledger"
	. "DNA/net/protocol"
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

	h.p.len = 1
	buf := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
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
	err := binary.Read(buf, binary.LittleEndian, &msg)
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

	for i := 0; i < int(msg.cnt); i++ {
		var headers ledger.Blockdata
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

func (msg headersReq) Handle(node Noder) error {
	log.Trace()
	// lock
	var starthash [HASHLEN]byte
	var stophash [HASHLEN]byte
	starthash = msg.p.hashStart
	stophash = msg.p.hashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt, err := GetHeadersFromHash(starthash, stophash)
	if err != nil {
		return err
	}
	buf, err := NewHeaders(headers, cnt)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}

func (msg blkHeader) Handle(node Noder) error {
	log.Trace()
	for i := 0; i < int(msg.cnt); i++ {
		var header ledger.Header
		header.Blockdata = &msg.blkHdr[i]
		err := ledger.DefaultLedger.Store.SaveHeader(&header, ledger.DefaultLedger)
		if err != nil {
			log.Warn("Add block Header error")
			return errors.New("Add block Header error\n")
		}
	}
	return nil
}
func GetHeadersFromHash(starthash common.Uint256, stophash common.Uint256) ([]ledger.Blockdata, uint32, error) {
	var count uint32 = 0
	var empty [HASHLEN]byte
	headers := []ledger.Blockdata{}
	var startheight uint32
	var stopheight uint32
	curHeight := ledger.DefaultLedger.GetLocalBlockChainHeight()
	if starthash == empty {
		if curHeight > MAXBLKHDRCNT {
			count = MAXBLKHDRCNT
		} else {
			count = curHeight
		}
	} else {
		bkstart, err := ledger.DefaultLedger.GetBlockWithHash(starthash)
		if err != nil {
			return nil, 0, err
		}
		startheight = bkstart.Blockdata.Height
		if stophash != empty {
			bkstop, err := ledger.DefaultLedger.GetBlockWithHash(stophash)
			if err != nil {
				return nil, 0, err
			}
			stopheight = bkstop.Blockdata.Height
			count = startheight - stopheight
			if count >= MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
				stopheight = startheight + MAXBLKHDRCNT
			}
		} else {

			if startheight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = startheight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		bk, err := ledger.DefaultLedger.GetBlockWithHeight(stopheight + i)
		if err != nil {
			log.Error("GetBlockWithHeight failed ", err.Error())
			return nil, 0, err
		}
		//log.Debug("GetHeadersFromHash height is ", i)
		//log.Debug("GetHeadersFromHash header is ", *bk.Blockdata)
		headers = append(headers, *bk.Blockdata)
	}

	return headers, count, nil
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
