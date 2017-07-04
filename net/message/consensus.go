package message

import (
	"DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/contract"
	"DNA/core/contract/program"
	"DNA/core/ledger"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"DNA/events"
	. "DNA/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookKeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           *crypto.PubKey
	Program         *program.Program

	hash common.Uint256
}

type consensus struct {
	msgHdr
	cons ConsensusPayload
	//event *events.Event
	//TBD
}

func (cp *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

func (cp *ConsensusPayload) Verify() error {
	return nil
}

func (cp *ConsensusPayload) InvertoryType() common.InventoryType {
	return common.CONSENSUS
}

func (cp *ConsensusPayload) GetProgramHashes() ([]common.Uint160, error) {
	log.Debug()

	if ledger.DefaultLedger == nil {
		return nil, errors.New("The Default ledger not exists.")
	}
	if cp.PrevHash != ledger.DefaultLedger.Store.GetCurrentBlockHash() {
		return nil, errors.New("The PreHash Not matched.")
	}

	contract, err := contract.CreateSignatureContract(cp.Owner)
	hash := contract.ProgramHash
	log.Debug("program hash == ", hash)

	//signatureRedeemScript, err := contract.CreateSignatureRedeemScript(bookKeepers[cp.BookKeeperIndex])
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Consensus], CreateSignatureContract failed.")
	}

	//hash, err:=common.ToCodeHash(signatureRedeemScript)
	//if err != nil {
	//	return  nil, NewDetailErr(err, ErrNoCode, "[Consensus], ToCodeHash failed.")
	//}
	programhashes := []common.Uint160{}
	programhashes = append(programhashes, hash)
	return programhashes, nil
}

func (cp *ConsensusPayload) SetPrograms(programs []*program.Program) {
	if programs == nil {
		log.Warn("Set programs with NULL parameters")
		return
	}

	if len(programs) > 0 {
		cp.Program = programs[0]
	} else {
		log.Warn("Set programs with 0 program")
	}
}

func (cp *ConsensusPayload) GetPrograms() []*program.Program {
	cpg := []*program.Program{}
	cpg = append(cpg, cp.Program)
	return cpg
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	return sig.GetHashData(cp)
	//return []byte{}
}

func (msg consensus) Handle(node Noder) error {
	log.Debug()
	node.LocalNode().GetEvent("consensus").Notify(events.EventNewInventory, &msg.cons)
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
func (cp *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

func (cp *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	serialization.WriteUint32(w, cp.Version)
	cp.PrevHash.Serialize(w)
	serialization.WriteUint32(w, cp.Height)
	serialization.WriteUint16(w, cp.BookKeeperIndex)
	serialization.WriteUint32(w, cp.Timestamp)
	serialization.WriteVarBytes(w, cp.Data)
	err := cp.Owner.Serialize(w)
	if err != nil {
		return err
	}
	return nil

}

func (cp *ConsensusPayload) Serialize(w io.Writer) error {
	err := cp.SerializeUnsigned(w)
	if cp.Program == nil {
		log.Error("Program is NULL")
		return errors.New("Program in consensus is NULL")
	}
	err = cp.Program.Serialize(w)
	return err
}

func (msg *consensus) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = msg.cons.Serialize(buf)

	return buf.Bytes(), err
}

func (cp *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	cp.Version, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Version Deserialize failed.")
		return errors.New("consensus item Version Deserialize failed.")
	}

	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		log.Warn("consensus item preHash Deserialize failed.")
		return errors.New("consensus item preHash Deserialize failed.")
	}
	cp.PrevHash = *preBlock

	cp.Height, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Height Deserialize failed.")
		return errors.New("consensus item Height Deserialize failed.")
	}

	cp.BookKeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {
		log.Warn("consensus item BookKeeperIndex Deserialize failed.")
		return errors.New("consensus item BookKeeperIndex Deserialize failed.")
	}

	cp.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Timestamp Deserialize failed.")
		return errors.New("consensus item Timestamp Deserialize failed.")
	}

	cp.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("consensus item Data Deserialize failed.")
		return errors.New("consensus item Data Deserialize failed.")
	}
	pk := new(crypto.PubKey)
	err = pk.DeSerialize(r)
	if err != nil {
		log.Warn("consensus item Owner deserialize failed.")
		return errors.New("consensus item Owner deserialize failed.")
	}
	cp.Owner = pk

	return nil
}

func (cp *ConsensusPayload) Deserialize(r io.Reader) error {
	err := cp.DeserializeUnsigned(r)

	pg := new(program.Program)
	err = pg.Deserialize(r)
	if err != nil {
		log.Error("Blockdata item Program Deserialize failed")
		return NewDetailErr(err, ErrNoCode, "Blockdata item Program Deserialize failed.")
	}
	cp.Program = pg
	return err
}

func (msg *consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	err = msg.cons.Deserialize(buf)
	return err
}

func NewConsensus(cp *ConsensusPayload) ([]byte, error) {
	log.Debug()
	var msg consensus
	msg.msgHdr.Magic = NETMAGIC
	cmd := "consensus"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	cp.Serialize(tmpBuffer)
	msg.cons = *cp
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
	log.Debug("NewConsensus The message payload length is ", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
