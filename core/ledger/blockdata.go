package ledger

import (
	"GoOnchain/common"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract/program"
	sig "GoOnchain/core/signature"
	//store "GoOnchain/core/store"
	. "GoOnchain/errors"
	"crypto/sha256"
	"errors"
	"io"
)

type Blockdata struct {
	Version          uint32
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData uint64
	NextMiner common.Uint160
	Program *program.Program

	hash common.Uint256
}

//Serialize the blockheader
func (bd *Blockdata) Serialize(w io.Writer) {
	bd.SerializeUnsigned(w)
	w.Write([]byte{byte(1)})
	bd.Program.Serialize(w)
}

//Serialize the blockheader data without program
func (bd *Blockdata) SerializeUnsigned(w io.Writer) error {
	//REVD: implement blockheader SerializeUnsigned
	serialization.WriteVarUint(w, uint64(bd.Version))
	bd.PrevBlockHash.Serialize(w)
	bd.TransactionsRoot.Serialize(w)
	serialization.WriteVarUint(w, uint64(bd.Timestamp))
	serialization.WriteVarUint(w, uint64(bd.Height))
	serialization.WriteVarUint(w, uint64(bd.ConsensusData))
	return nil
}

func (bd *Blockdata) Deserialize(r io.Reader) error {
	//REVD：Blockdata Deserialize
	bd.DeserializeUnsigned(r)
	p := make([]byte, 1)
	n, err := r.Read(p)
	if n > 0 {
		x := []byte(p[:])

		if x[0] != byte(1) {
			return NewDetailErr(errors.New("Blockdata Deserialize get format error."), ErrNoCode, "")
		}
	} else {
		return NewDetailErr(errors.New("Blockdata Deserialize get format error."), ErrNoCode, "")
	}
	pg := new(program.Program)
	err = pg.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item Program Deserialize failed.")
	}
	bd.Program = pg
	return nil
}

func (bd *Blockdata) DeserializeUnsigned(r io.Reader) error {
	//Version
	temp, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item Version Deserialize failed.")
	}
	bd.Version = uint32(temp)

	//PrevBlockHash
	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item preBlock Deserialize failed.")
	}
	bd.PrevBlockHash = *preBlock

	//TransactionsRoot
	txRoot := new(common.Uint256)
	err = txRoot.Deserialize(r)
	if err != nil {
		return err
	}
	bd.TransactionsRoot = *txRoot

	//Timestamp
	temp, _ = serialization.ReadVarUint(r, 0)
	bd.Timestamp = uint32(temp)

	//Height
	temp, _ = serialization.ReadVarUint(r, 0)
	bd.Height = uint32(temp)

	//consensusData
	bd.ConsensusData, _ = serialization.ReadVarUint(r, 0)

	return nil
}

func (bd *Blockdata) GetProgramHashes() ([]common.Uint160, error) {
	//TODO: implement blockheader GetProgramHashes

	//	if PrevBlockHash == new UInt256(){
	//		return bd.Program.CodeHash(),nil
	//	}
	//	prev_header, _:= store.ldbs.GetBlock(bd.PrevBlockHash)
	//	return new UInt160[] { prev_header.NextMiner };

	//	Blockchain.Default.GetHeader(PrevBlock);
	//	programHashes := []common.Uint160{}
	//	outputHashes, _ := bd.GetOutputHashes() //check error
	//	programHashes = append(programHashes, outputHashes[:]...)

	//	return programHashes, nil
	programHashes := []common.Uint160{}
	outputHashes := bd.Program.CodeHash()
	programHashes = append(programHashes, outputHashes)
	return programHashes, nil
}

func (bd *Blockdata) SetPrograms(programs []*program.Program) {
	if len(programs) != 1 {
		return
	}
	bd.Program = programs[0]
}

func (bd *Blockdata) GetPrograms() []*program.Program {
	return []*program.Program{bd.Program}
}

func (bd *Blockdata) Hash() common.Uint256 {
	//TODO: implement Blockdata Hash
	d := sig.GetHashData(bd)
	temp := sha256.Sum256([]byte(d))
	f := sha256.Sum256(temp[:])
	hash := common.Uint256(f)
	return hash
}

func  (bd *Blockdata) GetMessage() ([]byte){
	//TODO: implement GetMessage()
	return  sig.GetHashData(bd)
}
