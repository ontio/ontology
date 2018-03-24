package p2pserver

/*
func (msg addr) Handle(node Noder) error {
	log.Debug()
	for _, v := range msg.nodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		//address := ip.To4().String() + ":" + strconv.Itoa(int(v.Port))
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Info(fmt.Sprintf("The ip address is %s id is 0x%x", address, v.ID))

		if v.ID == node.LocalNode().GetID() {
			continue
		}
		if node.LocalNode().NodeEstablished(v.ID) {
			continue
		}

		if v.Port == 0 {
			continue
		}

		go node.LocalNode().Connect(address, false)
	}
	return nil
}
func (msg addrReq) Handle(node Noder) error {
	log.Debug()
	// lock
	var addrstr []NodeAddr
	var count uint64
	addrstr, count = node.LocalNode().GetNeighborAddrs()
	buf, err := NewAddrs(addrstr, count)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}
func (msg block) Handle(node Noder) error {
	log.Debug("RX block message")
	hash := msg.blk.Hash()
	if con, _ := actor.IsContainBlock(hash); con != true {
		actor.AddBlock(&msg.blk)
	} else {
		log.Debug("Receive duplicated block")
	}
	return nil
}

func (msg dataReq) Handle(node Noder) error {
	log.Debug()
	reqtype := common.InventoryType(msg.dataType)
	hash := msg.hash
	switch reqtype {
	case common.BLOCK:
		block, err := NewBlockFromHash(hash)
		if err != nil {
			log.Debug("Can't get block from hash: ", hash, " ,send not found message")
			//call notfound message
			b, err := NewNotFound(hash)
			node.Tx(b)
			return err
		}
		log.Debug("block height is ", block.Header.Height, " ,hash is ", hash)
		buf, err := NewBlock(block)
		if err != nil {
			return err
		}
		node.Tx(buf)

	case common.TRANSACTION:
		txn, err := NewTxnFromHash(hash)
		if err != nil {
			return err
		}
		buf, err := NewTxn(txn)
		if err != nil {
			return err
		}
		go node.Tx(buf)
	}
	return nil
}

func NewBlockFromHash(hash common.Uint256) (*types.Block, error) {
	//bk, err := ledger.DefaultLedger.Store.GetBlock(hash)
	bk, err := actor.GetBlockByHash(hash)
	if err != nil {
		log.Errorf("Get Block error: %s, block hash: %x", err.Error(), hash)
		return nil, err
	}
	return bk, nil
}
func (msg headersReq) Handle(node Noder) error {
	log.Debug()
	// lock
	node.LocalNode().AcqSyncReqSem()
	defer node.LocalNode().RelSyncReqSem()
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
		go node.Tx(buf)
	}
}

func (msg blkHeader) Handle(node Noder) error {
	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < msg.cnt; i++ {
		blkHdr = append(blkHdr, &msg.blkHdr[i])
	}
	actor.AddHeaders(blkHdr)
	return nil
}

func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]types.Header, uint32, error) {
	var count uint32 = 0
	var empty [HASHLEN]byte
	headers := []types.Header{}
	var startHeight uint32
	var stopHeight uint32
	//curHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	curHeight, _ := actor.GetCurrentHeaderHeight()
	if startHash == empty {
		if stopHash == empty {
			if curHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = curHeight
			}
		} else {
			//bkstop, err := ledger.DefaultLedger.Store.GetHeader(stopHash)
			bkstop, err := actor.GetHeaderByHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Height
			count = curHeight - stopHeight
			if count > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			}
		}
	} else {
		bkstart, err := actor.GetHeaderByHash(startHash)
		if err != nil {
			return nil, 0, err
		}
		startHeight = bkstart.Height
		if stopHash != empty {
			bkstop, err := actor.GetHeaderByHash(stopHash)
			if err != nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, 0, errors.New("do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
				stopHeight = startHeight - MAXBLKHDRCNT
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
		hash, err := actor.GetBlockHashByHeight(stopHeight + i)
		hd, err := actor.GetHeaderByHash(hash)
		if err != nil {
			log.Errorf("GetBlockWithHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		headers = append(headers, *hd)
	}

	return headers, count, nil
}
func (msg consensus) Handle(node Noder) error {
	log.Debug()
	//node.LocalNode().GetEvent("consensus").Notify(events.EventNewInventory, &msg.cons)
	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&msg.cons)
	}
	return nil
}
func (msg blocksReq) Handle(node Noder) error {
	log.Debug()
	log.Debug("handle blocks request")
	var starthash Uint256
	var stophash Uint256
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
func (msg Inv) Handle(node Noder) error {
	log.Debug()
	var id Uint256
	str := hex.EncodeToString(msg.P.Blk)
	log.Debug(fmt.Sprintf("The inv type: 0x%x block len: %d, %s\n",
		msg.P.InvType, len(msg.P.Blk), str))

	invType := InventoryType(msg.P.InvType)
	switch invType {
	case TRANSACTION:
		log.Debug("RX TRX message")
		// TODO check the ID queue
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			reqTxnData(node, id)
		}
	case BLOCK:
		log.Debug("RX block message")
		var i uint32
		count := msg.P.Cnt
		log.Debug("RX inv-block message, hash is ", msg.P.Blk)
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(msg.P.Blk[HASHLEN*i:]))
			// TODO check the ID queue
			//if !ledger.DefaultLedger.Store.BlockInCache(id) &&
			//	!ledger.DefaultLedger.BlockInLedger(id) &&
			//	LastInvHash != id {
			//	LastInvHash = id
			//	// send the block request
			//	log.Infof("inv request block hash: %x", id)
			//	ReqBlkData(node, id)
			//}
			isContainBlock, _ := actor.IsContainBlock(id)
			if !isContainBlock && LastInvHash != id {
				LastInvHash = id
				// send the block request
				log.Infof("inv request block hash: %x", id)
				ReqBlkData(node, id)
			}

		}
	case CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		reqConsensusData(node, id)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}
func GetInvFromBlockHash(starthash Uint256, stophash Uint256) (*InvPayload, error) {
	var count uint32 = 0
	var i uint32
	var empty Uint256
	var startheight uint32
	var stopheight uint32
	//curHeight := ledger.DefaultLedger.GetLocalBlockChainHeight()
	curHeight, _ := actor.GetCurrentBlockHeight()
	if starthash == empty {
		if stophash == empty {
			if curHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = curHeight
			}
		} else {
			//bkstop, err := ledger.DefaultLedger.Store.GetHeader(stophash)
			bkstop, err := actor.GetHeaderByHash(stophash)
			if err != nil {
				return nil, err
			}
			stopheight = bkstop.Height
			count = curHeight - stopheight
			if curHeight > MAXINVHDRCNT {
				count = MAXINVHDRCNT
			}
		}
	} else {
		//bkstart, err := ledger.DefaultLedger.Store.GetHeader(starthash)
		bkstart, err := actor.GetHeaderByHash(starthash)
		if err != nil {
			return nil, err
		}
		startheight = bkstart.Height
		if stophash != empty {
			//bkstop, err := ledger.DefaultLedger.Store.GetHeader(stophash)
			bkstop, err := actor.GetHeaderByHash(stophash)
			if err != nil {
				return nil, err
			}
			stopheight = bkstop.Height
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
		//hash, _ := ledger.DefaultLedger.Store.GetBlockHash(stopheight + i)
		hash, _ := actor.GetBlockHashByHeight(stopheight + i)
		log.Debug("GetInvFromBlockHash i is ", i, " , hash is ", hash)
		hash.Serialize(tmpBuffer)
	}
	log.Debug("GetInvFromBlockHash hash is ", tmpBuffer.Bytes())
	return NewInvPayload(BLOCK, count, tmpBuffer.Bytes()), nil
}

// FIXME the length exceed int32 case?
func HandleNodeMsg(node Noder, buf []byte, len int) error {
	if len < MSGHDRLEN {
		log.Warn("Unexpected size of received message")
		return errors.New("Unexpected size of received message")
	}

	log.Debugf("Received data len:  %d\n%x", len, buf[:len])

	s, err := MsgType(buf)
	if err != nil {
		log.Error("Message type parsing error")
		return err
	}

	msg := AllocMsg(s, len)
	if msg == nil {
		log.Error(fmt.Sprintf("Allocation message %s failed", s))
		return errors.New("Allocation message failed")
	}
	// Todo attach a node pointer to each message
	// Todo drop the message when verify/deseria packet error
	msg.Deserialization(buf[:len])
	msg.Verify(buf[MSGHDRLEN:len])

	return msg.Handle(node)
}


func (hdr msgHdr) Handle(n Noder) error {
	log.Debug()
	// TBD
	return nil
}
func (msg notFound) Handle(node Noder) error {
	log.Debug("RX notfound message, hash is ", msg.hash)
	return nil
}
func (msg pong) Handle(node Noder) error {
	node.SetHeight(msg.height)
	return nil
}
func (msg trn) Handle(node Noder) error {
	log.Debug()
	log.Debug("RX Transaction message")
	tx := &msg.txn
	if !node.LocalNode().ExistedID(tx.Hash()) {
		//if errCode := node.LocalNode().AppendTxnPool(&(msg.txn)); errCode != ErrNoError {
		//	return errors.New("[message] VerifyTransaction failed when AppendTxnPool.")
		//}
		actor.AddTransaction(&msg.txn)
		//node.LocalNode().IncRxTxnCnt()
		log.Debug("RX Transaction message hash", msg.txn.Hash())
		//log.Debug("RX Transaction message type", msg.txn.TxType)
	}

	return nil
}
func NewTxnFromHash(hash common.Uint256) (*types.Transaction, error) {
	//txn, err := ledger.DefaultLedger.GetTransactionWithHash(hash)
	txn, err := actor.GetTxnFromLedger(hash)
	if err != nil {
		log.Error("Get transaction with hash error: ", err.Error())
		return nil, err
	}
	return txn, nil
}

func (msg verACK) Handle(node Noder) error {
	log.Debug()

	if msg.isConsensus == true {
		s := node.GetConsensusState()
		if s != HANDSHAKE && s != HANDSHAKED {
			log.Warn("Unknow status to received verack")
			return errors.New("Unknow status to received verack")
		}

		localNode := node.LocalNode()
		n, ok := localNode.GetNbrNode(node.GetID())
		if ok == false {
			log.Warn("nbr node is not exsit")
			return errors.New("nbr node is not exsit")
		}

		node.SetConsensusState(ESTABLISH)
		n.SetConsensusState(node.GetConsensusState())
		n.SetConsensusConn(node.GetConsensusConn())
		//	n.SetConsensusPort(node.GetConsensusPort())
		//	n.SetConsensusState(node.GetConsensusState())

		if s == HANDSHAKE {
			buf, _ := NewVerack(true)
			node.ConsensusTx(buf)
		}
		return nil
	}
	s := node.GetState()
	if s != HANDSHAKE && s != HANDSHAKED {
		log.Warn("Unknow status to received verack")
		return errors.New("Unknow status to received verack")
	}

	node.SetState(ESTABLISH)

	if s == HANDSHAKE {
		buf, _ := NewVerack(false)
		node.Tx(buf)
	}

	node.DumpInfo()
	// Fixme, there is a race condition here,
	// but it doesn't matter to access the invalid
	// node which will trigger a warning
	//TODO JQ: only master p2p port request neighbor list
	node.ReqNeighborList()
	addr := node.GetAddr()
	port := node.GetPort()
	nodeAddr := addr + ":" + strconv.Itoa(int(port))
	//TODO JQ： only master p2p port remove the list
	node.LocalNode().RemoveAddrInConnectingList(nodeAddr)
	//connect consensus port

	if s == HANDSHAKED {
		consensusPort := node.GetConsensusPort()
		nodeConsensusAddr := addr + ":" + strconv.Itoa(int(consensusPort))
		go node.Connect(nodeConsensusAddr, true)
	}
	return nil
}

func (msg version) Handle(node Noder) error {
	log.Debug()
	localNode := node.LocalNode()

	// Exclude the node itself
	if msg.P.Nonce == localNode.GetID() {
		if msg.P.IsConsensus == false {
			log.Warn("The node handshark with itself")
			node.CloseConn()
			return errors.New("The node handshark with itself")
		}
		if msg.P.IsConsensus == true {
			log.Warn("The node handshark with itself")
			node.CloseConsensusConn()
			return errors.New("The node handshark with itself")
		}
	}

	if msg.P.IsConsensus == true {
		s := node.GetConsensusState()
		if s != INIT && s != HAND {
			log.Warn("Unknow status to received version")
			return errors.New("Unknow status to received version")
		}

		//	n, ok := LocalNode.GetNbrNode(msg.P.Nonce)
		//	if ok == false {
		//		log.Warn("nbr node is not exsit")
		//		return errors.New("nbr node is not exsit")
		//	}

		//	n.SetConsensusConn(node.GetConsensusConn())
		//	n.SetConsensusPort(node.GetConsensusPort())
		//	n.SetConsensusState(node.GetConsensusState())

		node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
			msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
		node.SetConsensusPort(msg.P.ConsensusPort)

		var buf []byte
		if s == INIT {
			node.SetConsensusState(HANDSHAKE)
			buf, _ = NewVersion(localNode, true)
		} else if s == HAND {
			node.SetConsensusState(HANDSHAKED)
			buf, _ = NewVerack(true)
		}
		node.ConsensusTx(buf)
		return nil
	}

	s := node.GetState()
	if s != INIT && s != HAND {
		log.Warn("Unknow status to received version")
		return errors.New("Unknow status to received version")
	}

	// Obsolete node
	n, ret := localNode.DelNbrNode(msg.P.Nonce)
	if ret == true {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", msg.P.Nonce))
		// Close the connection and release the node soure
		n.SetState(INACTIVITY)
		n.CloseConn()
	}

	log.Debug("handle version msg.pk is ", msg.pk)
	if msg.P.Cap[HTTPINFOFLAG] == 0x01 {
		node.SetHttpInfoState(true)
	} else {
		node.SetHttpInfoState(false)
	}
	node.SetHttpInfoPort(msg.P.HttpInfoPort)
	node.SetConsensusPort(msg.P.ConsensusPort)
	node.SetBookKeeperAddr(msg.pk)
	// if  msg.P.Port == msg.P.ConsensusPort don't updateInfo
	node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
		msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
	localNode.AddNbrNode(node)

	var buf []byte
	if s == INIT {
		node.SetState(HANDSHAKE)
		buf, _ = NewVersion(localNode, false)
	} else if s == HAND {
		node.SetState(HANDSHAKED)
		buf, _ = NewVerack(false)
	}
	node.Tx(buf)

	return nil
}
func ReqBlkData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.BLOCK
	msg.hash = hash

	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "getdata")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.dataType))
	msg.hash.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new getdata Msg")
		return err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.msgHdr.Length)

	sendBuf, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return err
	}

	node.Tx(sendBuf)

	return nil
}
func reqConsensusData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.CONSENSUS
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
func NewHeadersReq() ([]byte, error) {
	var h headersReq

	h.p.len = 1
	//buf := ledger.DefaultLedger.Store.GetCurrentHeaderHash()
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
	return []byte{}, nil
}
func NewBlocksReq(n Noder) ([]byte, error) {
	var h blocksReq
	log.Debug("request block hash")
	// Fixme correct with the exactly request length
	h.p.HeaderHashCount = 1
	//Fixme! Should get the remote Node height.
	//buf := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	buf, _ := actor.GetCurrentBlockHash()

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
// TODO combine all of message alloc in one function via interface
func NewMsg(t string, n Noder) ([]byte, error) {
	switch t {
	case "version":
		return NewVersion(n, false)
	case "verack":
		return NewVerack(false)
	case "getheaders":
		return NewHeadersReq()
	case "getaddr":
		return newGetAddr()

	default:
		return nil, fmt.Errorf("Unknown message type %v", t)
	}
}
func NewPingMsg() ([]byte, error) {
	var msg ping
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "ping")
	//msg.height = uint64(ledger.DefaultLedger.Blockchain.BlockHeight)
	height, _ := actor.GetCurrentBlockHeight()
	msg.height = uint64(height)
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.height)
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
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
func (msg ping) Handle(node Noder) error {
	node.SetHeight(msg.height)
	buf, err := NewPongMsg()
	if err != nil {
		log.Error("failed build a new ping message")
	} else {
		go node.Tx(buf)
	}
	return err
}
func NewPongMsg() ([]byte, error) {
	var msg pong
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "pong")
	//msg.height = uint64(ledger.DefaultLedger.Store.GetHeaderHeight())
	height, _ := actor.GetCurrentHeaderHeight()
	msg.height = uint64(height)
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.height)
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
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
func reqTxnData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.TRANSACTION
	// TODO handle the hash array case
	//msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)
	return nil
}
func ReqTxnPool(node Noder) error {
	msg := AllocMsg("txnpool", 0)
	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
func (msg *version) init(n Noder) {
	// Do the init
}

func NewVersion(n Noder, isConsensus bool) ([]byte, error) {
	log.Debug()
	var msg version

	msg.P.Version = n.Version()
	msg.P.Services = n.Services()
	msg.P.HttpInfoPort = config.Parameters.HttpInfoPort
	msg.P.ConsensusPort = n.GetConsensusPort()
	msg.P.IsConsensus = isConsensus
	if config.Parameters.HttpInfoStart {
		msg.P.Cap[HTTPINFOFLAG] = 0x01
	} else {
		msg.P.Cap[HTTPINFOFLAG] = 0x00
	}

	// FIXME Time overflow
	msg.P.TimeStamp = uint32(time.Now().UTC().UnixNano())
	msg.P.Port = n.GetPort()
	msg.P.Nonce = n.GetID()
	msg.P.UserAgent = 0x00
	//msg.P.StartHeight = 0 //uint64(ledger.DefaultLedger.GetLocalBlockChainHeight())
	height, _ := actor.GetCurrentBlockHeight()
	msg.P.StartHeight = uint64(height)
	if n.GetRelay() {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}

	msg.pk = n.GetBookKeeperAddr()
	log.Debug("new version msg.pk is ", msg.pk)
	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	msg.Hdr.Magic = NETMAGIC
	copy(msg.Hdr.CMD[0:7], "version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	msg.pk.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.Hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}
*/
