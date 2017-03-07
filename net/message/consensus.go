package message

import (
	"GoOnchain/common"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract/program"
	//"GoOnchain/events"
	. "GoOnchain/net/protocol"
	"bytes"
	"GoOnchain/events"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"GoOnchain/core/contract"
	"GoOnchain/core/ledger"
	"strconv"
	. "GoOnchain/errors"
	sig "GoOnchain/core/signature"
)

type ConsensusPayload struct {
	Version    uint32
	PrevHash   common.Uint256
	Height     uint32
	MinerIndex uint16
	Timestamp  uint32
	Data       []byte
	Program    *program.Program

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
	common.Trace()
	if ledger.DefaultLedger == nil{
		return nil,errors.New("The Default ledger not exists.")
	}
	if cp.PrevHash != ledger.DefaultLedger.Store.GetCurrentBlockHash(){
		return nil,errors.New("The PreHash Not matched.")
	}
	miners := ledger.DefaultLedger.Blockchain.GetMiners()
	if uint16(len(miners)) <= cp.MinerIndex{
		return nil,errors.New("MinerIndex invalidate. miners lengths="+ strconv.Itoa(len(miners))+ "MinerIndex ="+strconv.Itoa(int(cp.MinerIndex)))
	}
	contract,err:= contract.CreateSignatureContract(miners[cp.MinerIndex])
	hash := contract.ProgramHash
	fmt.Println("program hash== ",hash)

	//signatureRedeemScript, err := contract.CreateSignatureRedeemScript(miners[cp.MinerIndex])
	if err != nil {
		return  nil, NewDetailErr(err, ErrNoCode, "[Consensus], CreateSignatureContract failed.")
	}

	//hash, err:=common.ToCodeHash(signatureRedeemScript)
	//if err != nil {
	//	return  nil, NewDetailErr(err, ErrNoCode, "[Consensus], ToCodeHash failed.")
	//}
	programhashes := []common.Uint160{}
	programhashes = append(programhashes,hash)
	return programhashes, nil
}

func (cp *ConsensusPayload) SetPrograms(programs []*program.Program) {
	cp.Program = programs[0]
}

func (cp *ConsensusPayload) GetPrograms() []*program.Program {
	cpg := []*program.Program{}
	cpg = append(cpg,cp.Program)
	return cpg
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	return sig.GetHashForSigning(cp)
	//return []byte{}
}

func (msg consensus) Handle(node Noder) error {
	common.Trace()

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
	serialization.WriteUint16(w, cp.MinerIndex)
	serialization.WriteUint32(w, cp.Timestamp)
	err := serialization.WriteVarBytes(w, cp.Data)
	if err != nil {
		return err
	}
	return nil

}

func (cp *ConsensusPayload) Serialize(w io.Writer) error {
	err := cp.SerializeUnsigned(w)
	if cp.Program == nil {
		common.Trace()
		fmt.Println("Program is NULL")
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
	common.Trace()
	var err error
	cp.Version, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("consensus item Version Deserialize failed.")
	}

	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		return errors.New("consensus item preHash Deserialize failed.")
	}
	cp.PrevHash = *preBlock

	cp.Height, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("consensus item Height Deserialize failed.")
	}

	cp.MinerIndex, err = serialization.ReadUint16(r)
	if err != nil {
		return errors.New("consensus item MinerIndex Deserialize failed.")
	}

	cp.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		return errors.New("consensus item Timestamp Deserialize failed.")
	}

	cp.Data, err = serialization.ReadVarBytes(r)
	fmt.Printf("The consensus payload data len is %d\n", len(cp.Data))
	if err != nil {
		return errors.New("consensus item Data Deserialize failed.")
	}
	return nil
}

func (cp *ConsensusPayload) Deserialize(r io.Reader) error {
	common.Trace()
	err := cp.DeserializeUnsigned(r)

	pg := new(program.Program)
	err = pg.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item Program Deserialize failed.")
	}
	cp.Program = pg
	return err
}

func (msg *consensus) Deserialization(p []byte) error {
	common.Trace()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	err = msg.cons.Deserialize(buf)
	return err
}

func NewConsensus(cp *ConsensusPayload) ([]byte, error) {
	common.Trace()
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
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))
	fmt.Printf("NewConsensus The message payload length is %d\n", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)

	return m, nil
}
