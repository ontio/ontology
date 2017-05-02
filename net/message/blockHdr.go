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
	blkHdr []ledger.Header
}

func NewHeadersReq() ([]byte, error) {
	var h headersReq

	h.p.len = 1
	buf := ledger.DefaultLedger.Store.GetCurrentHeaderHash()
	copy(h.p.hashEnd[:], reverse(buf[:]))

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
	binary.Write(buf, binary.LittleEndian, msg.cnt)

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
		var headers ledger.Header
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
	log.Debug()
	// lock
	var startHash [HASHLEN]byte
	var stopHash [HASHLEN]byte
	startHash = msg.p.hashStart
	stopHash = msg.p.hashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt, err := GetHeadersFromHash(startHash, stopHash)
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

func SendMsgSyncHeaders(node Noder) {
	buf, err := NewHeadersReq()
	if err != nil {
		log.Error("failed build a new headersReq")
	} else {
		node.LocalNode().SetSyncHeaders(true)
		node.SetSyncHeaders(true)
		go node.Tx(buf)
	}
}

func ReqBlkHdrFromOthers(node Noder) {
	node.SetSyncFailed()
	noders := node.LocalNode().GetNeighborNoder()
	for _, noder := range noders {
		if noder.IsSyncFailed() != true {
			SendMsgSyncHeaders(noder)
			break
		}
	}
}

func (msg blkHeader) Handle(node Noder) error {
	log.Debug()
	node.StopRetryTimer()
	err := ledger.DefaultLedger.Store.AddHeaders(msg.blkHdr, ledger.DefaultLedger)
	if err != nil {
		log.Warn("Add block Header error")
		ReqBlkHdrFromOthers(node)
		return errors.New("Add block Header error, send new header request to another node\n")
	}
	if msg.cnt == MAXBLKHDRCNT {
		SendMsgSyncHeaders(node)
		node.StartRetryTimer()
	}
	return nil
}

func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]ledger.Header, uint32, error) {
	var count uint32 = 0
	var empty [HASHLEN]byte
	headers := []ledger.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := ledger.DefaultLedger.GetLocalBlockChainHeight()
	if startHash == empty {
		if stopHash == empty {
			if curHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = curHeight
			}
		} else {
			bkstop, err := ledger.DefaultLedger.GetBlockWithHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Blockdata.Height
			count = curHeight - stopHeight
			if curHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			}
		}
	} else {
		bkstart, err := ledger.DefaultLedger.GetBlockWithHash(startHash)
		if err != nil {
			return nil, 0, err
		}
		startHeight = bkstart.Blockdata.Height
		if stopHash != empty {
			bkstop, err := ledger.DefaultLedger.GetBlockWithHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Blockdata.Height
			count = startHeight - stopHeight
			if count >= MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
				stopHeight = startHeight + MAXBLKHDRCNT
			}
		} else {

			if startHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash, err := ledger.DefaultLedger.Store.GetBlockHash(stopHeight + i)
		hd, err := ledger.DefaultLedger.Store.GetHeader(hash)
		if err != nil {
			log.Error("GetBlockWithHeight failed ", err.Error())
			return nil, 0, err
		}
		headers = append(headers, *hd)
	}

	return headers, count, nil
}

func NewHeaders(headers []ledger.Header, count uint32) ([]byte, error) {
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
