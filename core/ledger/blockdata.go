package ledger

import (
	. "DNA/common"
	"DNA/common/serialization"
	"DNA/core/contract/program"
	sig "DNA/core/signature"
	. "DNA/errors"
	"crypto/sha256"
	"errors"
	"io"
)

type Blockdata struct {
	Version          uint32
	PrevBlockHash    Uint256
	TransactionsRoot Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	NextBookKeeper   Uint160
	Program          *program.Program

	hash Uint256
}

//Serialize the blockheader
func (bd *Blockdata) Serialize(w io.Writer) {
	bd.SerializeUnsigned(w)
	w.Write([]byte{byte(1)})
	if bd.Program != nil {
		bd.Program.Serialize(w)
	}
}

//Serialize the blockheader data without program
func (bd *Blockdata) SerializeUnsigned(w io.Writer) error {
	//REVD: implement blockheader SerializeUnsigned
	serialization.WriteUint32(w, bd.Version)
	bd.PrevBlockHash.Serialize(w)
	bd.TransactionsRoot.Serialize(w)
	serialization.WriteUint32(w, bd.Timestamp)
	serialization.WriteUint32(w, bd.Height)
	serialization.WriteUint64(w, bd.ConsensusData)
	bd.NextBookKeeper.Serialize(w)
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
	temp, err := serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item Version Deserialize failed.")
	}
	bd.Version = temp

	//PrevBlockHash
	preBlock := new(Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Blockdata item preBlock Deserialize failed.")
	}
	bd.PrevBlockHash = *preBlock

	//TransactionsRoot
	txRoot := new(Uint256)
	err = txRoot.Deserialize(r)
	if err != nil {
		return err
	}
	bd.TransactionsRoot = *txRoot

	//Timestamp
	temp, _ = serialization.ReadUint32(r)
	bd.Timestamp = uint32(temp)

	//Height
	temp, _ = serialization.ReadUint32(r)
	bd.Height = uint32(temp)

	//consensusData
	bd.ConsensusData, _ = serialization.ReadUint64(r)

	//NextBookKeeper
	bd.NextBookKeeper.Deserialize(r)

	return nil
}

func (bd *Blockdata) GetProgramHashes() ([]Uint160, error) {
	programHashes := []Uint160{}
	zero := Uint256{}

	if bd.PrevBlockHash == zero {
		pg := *bd.Program
		outputHashes, err := ToCodeHash(pg.Code)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Blockdata], GetProgramHashes failed.")
		}
		programHashes = append(programHashes, outputHashes)
		return programHashes, nil
	} else {
		prev_header, err := DefaultLedger.Store.GetHeader(bd.PrevBlockHash)
		if err != nil {
			return programHashes, err
		}
		programHashes = append(programHashes, prev_header.Blockdata.NextBookKeeper)
		return programHashes, nil
	}

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

func (bd *Blockdata) Hash() Uint256 {

	d := sig.GetHashData(bd)
	temp := sha256.Sum256([]byte(d))
	f := sha256.Sum256(temp[:])
	hash := Uint256(f)
	return hash
}

func (bd *Blockdata) GetMessage() []byte {
	//return sig.GetHashData(bd)
	return sig.GetHashForSigning(bd)
}
