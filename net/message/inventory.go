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
	"encoding/hex"
	"fmt"
	"io"
)

type InventoryType byte

type blocksReq struct {
	msgHdr
	p struct {
		HeaderHashCount uint8
		hashStart       [HASHLEN]byte
		hashStop        [HASHLEN]byte
	}
}

type invPayload struct {
	InvType uint8
	Blk     []byte
}

type Inv struct {
	Hdr msgHdr
	P   invPayload
}

func NewBlocksReq(n Noder) ([]byte, error) {
	var h blocksReq
	log.Debug("request block hash")
	// Fixme correct with the exactly request length
	h.p.HeaderHashCount = 1
	//Fixme! Should get the remote Node height.
	buf := ledger.DefaultLedger.Blockchain.CurrentBlockHash()

	copy(h.p.hashStart[:], reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.msgHdr.init("getblocks", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()

	return m, err
}

func (msg blocksReq) Verify(buf []byte) error {

	// TODO verify the message Content
	err := msg.msgHdr.Verify(buf)
	return err
}

func (msg blocksReq) Handle(node Noder) error {
	common.Trace()
	log.Debug("handle blocks request")
	var starthash common.Uint256
	var stophash common.Uint256
	starthash = msg.p.hashStart
	stophash = msg.p.hashStop
	//FIXME if HeaderHashCount > 1
	inv, err := GetInvFromBlockHash(starthash, stophash)
	if err != nil {
		return err
	}
	buf, err := NewInv(inv)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}

func (msg blocksReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blocksReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &msg)
	return err
}

func (msg Inv) Verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Inv) Handle(node Noder) error {
	common.Trace()
	var id common.Uint256
	str := hex.EncodeToString(msg.P.Blk)
	fmt.Printf("The inv type: 0x%x block len: %d, %s\n",
		msg.P.InvType, len(msg.P.Blk), str)

	invType := common.InventoryType(msg.P.InvType)
	switch invType {
	case common.TRANSACTION:
		log.Debug("RX TRX message")
		// TODO check the ID queue
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			reqTxnData(node, id)
		}
	case common.BLOCK:
		log.Debug("RX block message")
		var i int
		count := len(msg.P.Blk) >> DIVHASHLEN
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(msg.P.Blk[HASHLEN*i+1:]))
			if !node.ExistedID(id) {
				// send the block request
				log.Info("inv request block hash: ", id)
				reqBlkData(node, id)
			}

		}
	case common.CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		reqConsensusData(node, id)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
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

	msg.P.InvType = p[MSGHDRLEN]
	msg.P.Blk = p[MSGHDRLEN+1:]
	return err
}

func (msg Inv) invType() byte {
	return msg.P.InvType
}

//func (msg inv) invLen() (uint64, uint8) {
func (msg Inv) invLen() (uint64, uint8) {
	var val uint64
	var size uint8

	len := binary.LittleEndian.Uint64(msg.P.Blk[0:1])
	if len < 0xfd {
		val = len
		size = 1
	} else if len == 0xfd {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:3])
		size = 3
	} else if len == 0xfe {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:5])
		size = 5
	} else if len == 0xff {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:9])
		size = 9
	}

	return val, size
}

func GetInvFromBlockHash(starthash common.Uint256, stophash common.Uint256) (invPayload, error) {
	var inv invPayload
	var count uint32 = 0
	var i uint32

	var empty common.Uint256
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
			return inv, err
		}
		startheight = bkstart.Blockdata.Height
		if stophash != empty {
			bkstop, err := ledger.DefaultLedger.GetBlockWithHash(stophash)
			if err != nil {
				return inv, err
			}
			stopheight = bkstop.Blockdata.Height
			count = startheight - stopheight
			if count >= MAXINVHDRCNT {
				count = MAXINVHDRCNT
				stopheight = startheight + MAXINVHDRCNT
			}
		} else {

			if startheight > MAXINVHDRCNT {
				count = MAXINVHDRCNT
			} else {
				count = startheight
			}
		}
	}
	tmpBuffer := bytes.NewBuffer([]byte{})
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHash
		hash, _ := ledger.DefaultLedger.Store.GetBlockHash(stopheight + i)
		hash.Serialize(tmpBuffer)
	}

	inv.Blk = tmpBuffer.Bytes()
	inv.InvType = 0x02
	return inv, nil
}

func NewInv(inv invPayload) ([]byte, error) {
	var msg Inv

	msg.P.Blk = inv.Blk
	msg.P.InvType = inv.InvType
	msg.Hdr.Magic = NETMAGIC
	cmd := "inv"
	copy(msg.Hdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	inv.Serialization(tmpBuffer)

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg", err.Error())
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg *invPayload) Serialization(w io.Writer) {
	serialization.WriteUint8(w, msg.InvType)
	serialization.WriteVarBytes(w, msg.Blk)
}
