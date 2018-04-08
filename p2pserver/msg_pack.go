package p2pserver

import (
	"time"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
)

func NewAddrs(nodeAddrs []msgCommon.PeerAddr, count uint64) ([]byte, error) {
	var addr msg.Addr
	addr.NodeAddrs = nodeAddrs
	addr.NodeCnt = count
	addr.Hdr.Magic = msgCommon.NETMAGIC
	cmd := "addr"
	copy(addr.Hdr.CMD[0:7], cmd)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, addr.NodeCnt)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, addr.NodeAddrs)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(addr.Hdr.Checksum))
	addr.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", addr.Hdr.Length)

	m, err := addr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewAddrReq() ([]byte, error) {
	var msg msg.AddrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.Init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message get addr length is: ", len(buf), " ", str)
	return buf, err
}

func NewBlock(bk *types.Block) ([]byte, error) {
	log.Debug()
	var blk msg.Block
	blk.Blk = *bk
	blk.MsgHdr.Magic = msgCommon.NETMAGIC
	cmd := "block"
	copy(blk.MsgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(blk.MsgHdr.Checksum))
	blk.MsgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", blk.MsgHdr.Length)

	m, err := blk.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewHeaders(headers []types.Header, count uint32) ([]byte, error) {
	var blkHdr msg.BlkHeader
	blkHdr.Cnt = count
	blkHdr.BlkHdr = headers
	blkHdr.Hdr.Magic = msgCommon.NETMAGIC
	cmd := "headers"
	copy(blkHdr.Hdr.CMD[0:len(cmd)], cmd)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, blkHdr.Cnt)
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
	binary.Read(buf, binary.LittleEndian, &(blkHdr.Hdr.Checksum))
	blkHdr.Hdr.Length = uint32(len(b.Bytes()))

	m, err := blkHdr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewHeadersReq(curHdrHash common.Uint256) ([]byte, error) {
	var h msg.HeadersReq

	h.P.Len = 1
	buf := curHdrHash
	copy(h.P.HashEnd[:], buf[:])

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new headersReq")
		return nil, err
	}

	s := msg.CheckSum(p.Bytes())
	h.Hdr.Init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
}

func NewBlocksReq(curBlkHash common.Uint256) ([]byte, error) {
	log.Debug("request block hash")
	var h msg.BlocksReq
	// Fixme correct with the exactly request length
	h.P.HeaderHashCount = 1
	//Fixme! Should get the remote Node height.

	buf := curBlkHash
	copy(h.P.HashStart[:], msg.Reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := msg.CheckSum(p.Bytes())
	h.MsgHdr.Init("getblocks", s, uint32(len(p.Bytes())))
	m, err := h.Serialization()
	return m, err
}

func NewConsensus(cp *msg.ConsensusPayload) ([]byte, error) {
	log.Debug()
	var cons msg.Consensus
	cons.MsgHdr.Magic = msgCommon.NETMAGIC
	cmd := "consensus"
	copy(cons.MsgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	cp.Serialize(tmpBuffer)
	cons.Cons = *cp
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
	binary.Read(buf, binary.LittleEndian, &(cons.MsgHdr.Checksum))
	cons.MsgHdr.Length = uint32(len(b.Bytes()))
	log.Debug("NewConsensus The message payload length is ", cons.MsgHdr.Length)

	m, err := cons.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewInv(invPayload *msg.InvPayload) ([]byte, error) {
	var inv msg.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType
	inv.P.Cnt = invPayload.Cnt
	inv.Hdr.Magic = msgCommon.NETMAGIC
	cmd := "inv"
	copy(inv.Hdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	invPayload.Serialization(tmpBuffer)

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
	binary.Read(buf, binary.LittleEndian, &(inv.Hdr.Checksum))
	inv.Hdr.Length = uint32(len(b.Bytes()))

	m, err := inv.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewNotFound(hash common.Uint256) ([]byte, error) {
	log.Debug()
	var notFound msg.NotFound
	notFound.Hash = hash
	notFound.MsgHdr.Magic = msgCommon.NETMAGIC
	cmd := "notfound"
	copy(notFound.MsgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	notFound.Hash.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new notfound Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(notFound.MsgHdr.Checksum))
	notFound.MsgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", notFound.MsgHdr.Length)

	m, err := notFound.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewPingMsg(height uint64) ([]byte, error) {
	var ping msg.Ping
	ping.Hdr.Magic = msgCommon.NETMAGIC
	copy(ping.Hdr.CMD[0:7], "ping")
	ping.Height = uint64(height)
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, ping.Height)
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
	binary.Read(buf, binary.LittleEndian, &(ping.Hdr.Checksum))
	ping.Hdr.Length = uint32(len(b.Bytes()))

	m, err := ping.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewPongMsg(height uint64) ([]byte, error) {
	var pong msg.Pong
	pong.MsgHdr.Magic = msgCommon.NETMAGIC
	copy(pong.MsgHdr.CMD[0:7], "pong")
	pong.Height = uint64(height)
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, pong.Height)
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
	binary.Read(buf, binary.LittleEndian, &(pong.MsgHdr.Checksum))
	pong.MsgHdr.Length = uint32(len(b.Bytes()))

	m, err := pong.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewTxn(txn *types.Transaction) ([]byte, error) {
	log.Debug()
	var trn msg.Trn

	trn.MsgHdr.Magic = msgCommon.NETMAGIC
	cmd := "tx"
	copy(trn.MsgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	txn.Serialize(tmpBuffer)
	trn.Txn = *txn
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
	binary.Read(buf, binary.LittleEndian, &(trn.MsgHdr.Checksum))
	trn.MsgHdr.Length = uint32(len(b.Bytes()))
	log.Debug("The message payload length is ", trn.MsgHdr.Length)

	m, err := trn.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewVerAck(isConsensus bool) ([]byte, error) {
	var verAck msg.VerACK
	verAck.MsgHdr.Magic = msgCommon.NETMAGIC
	copy(verAck.MsgHdr.CMD[0:7], "verack")
	verAck.IsConsensus = isConsensus
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, verAck.IsConsensus)
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
	binary.Read(buf, binary.LittleEndian, &(verAck.MsgHdr.Checksum))
	verAck.MsgHdr.Length = uint32(len(b.Bytes()))

	m, err := verAck.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewVersionPayload(version uint32, service uint64, port uint16,
	consPort uint16,  nonce uint64, startHeight uint64,
	relay bool, isCons bool) msg.VersionPayload {
	vpl :=  msg.VersionPayload{
		Version:       version,
		Services:      service,
		Port:          port,
		ConsensusPort: consPort,
		Nonce:         nonce,
		StartHeight:   startHeight,
		IsConsensus:   isCons,
	}
	if relay {
		vpl.Relay = 1
	} else {
		vpl.Relay = 0
	}
	if config.Parameters.HttpInfoStart {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x01
	} else {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x00
	}

	vpl.UserAgent = 0x00
	vpl.HttpInfoPort = config.Parameters.HttpInfoPort
	vpl.TimeStamp = uint32(time.Now().UTC().UnixNano())

	return vpl
}

func NewVersion(vpl msg.VersionPayload, pk *crypto.PubKey) ([]byte, error) {
	log.Debug()
	var version msg.Version
	version.P = vpl
	version.PK = pk
	log.Debug("new version msg.pk is ", version.PK)
	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	version.Hdr.Magic = msgCommon.NETMAGIC
	copy(version.Hdr.CMD[0:7], "version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(version.P))
	version.PK.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(version.Hdr.Checksum))
	version.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", version.Hdr.Length)

	m, err := version.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewTxnDataReq(hash common.Uint256)  ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.TRANSACTION
	// TODO handle the hash array case
	dataReq.Hash = hash

	buf, _ := dataReq.Serialization()
	return buf, nil
}

func NewBlkDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	dataReq.MsgHdr.Magic = msgCommon.NETMAGIC
	copy(dataReq.MsgHdr.CMD[0:7], "getdata")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(dataReq.DataType))
	dataReq.Hash.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new getdata Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(dataReq.MsgHdr.Checksum))
	dataReq.MsgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", dataReq.MsgHdr.Length)

	sendBuf, err := dataReq.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return sendBuf, nil
}

func NewConsensusDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.CONSENSUS
	// TODO handle the hash array case
	dataReq.Hash = hash
	buf, _ := dataReq.Serialization()
	return buf, nil
}

