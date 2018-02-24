package types

import (
	"bytes"
	"crypto/sha256"
	"io"

	. "github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/contract/program"
	. "github.com/Ontology/errors"
)

type Header struct {
	Version          uint32
	PrevBlockHash    Uint256
	TransactionsRoot Uint256
	StateRoot        Uint256
	BlockRoot        Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	NextBookKeeper   Address

	Program *program.Program

	hash Uint256
}

//Serialize the blockheader
func (bd *Header) Serialize(w io.Writer) {
	bd.SerializeUnsigned(w)
	w.Write([]byte{byte(1)})

	//TODO: fix this, inconsist with Deserialize
	if bd.Program != nil {
		bd.Program.Serialize(w)
	}
}

//Serialize the blockheader data without program
func (bd *Header) SerializeUnsigned(w io.Writer) error {
	serialization.WriteUint32(w, bd.Version)
	bd.PrevBlockHash.Serialize(w)
	bd.TransactionsRoot.Serialize(w)
	bd.StateRoot.Serialize(w)
	bd.BlockRoot.Serialize(w)
	serialization.WriteUint32(w, bd.Timestamp)
	serialization.WriteUint32(w, bd.Height)
	serialization.WriteUint64(w, bd.ConsensusData)
	bd.NextBookKeeper.Serialize(w)
	return nil
}

func (bd *Header) Deserialize(r io.Reader) error {
	bd.DeserializeUnsigned(r)

	pg := new(program.Program)
	err := pg.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Header item Program Deserialize failed.")
	}
	bd.Program = pg
	return nil
}

func (bd *Header) DeserializeUnsigned(r io.Reader) error {
	//Version
	temp, err := serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Header item Version Deserialize failed.")
	}
	bd.Version = temp

	//PrevBlockHash
	err = bd.PrevBlockHash.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Header item preBlock Deserialize failed.")
	}

	//TransactionsRoot
	err = bd.TransactionsRoot.Deserialize(r)
	if err != nil {
		return err
	}

	//StateRoot
	err = bd.StateRoot.Deserialize(r)
	if err != nil {
		return err
	}
	err = bd.BlockRoot.Deserialize(r)
	if err != nil {
		return err
	}

	//Timestamp
	temp, _ = serialization.ReadUint32(r)
	bd.Timestamp = uint32(temp)

	//Height
	temp, _ = serialization.ReadUint32(r)
	bd.Height = uint32(temp)

	//consensusData
	bd.ConsensusData, _ = serialization.ReadUint64(r)

	//NextBookKeeper
	err = bd.NextBookKeeper.Deserialize(r)

	return err
}

func (bd *Header) Hash() Uint256 {
	buf := new(bytes.Buffer)
	bd.SerializeUnsigned(buf)
	temp := sha256.Sum256(buf.Bytes())
	hash := sha256.Sum256(temp[:])
	return hash
}

func (bd *Header) GetMessage() []byte {
	bf := new(bytes.Buffer)
	bd.SerializeUnsigned(bf)
	return bf.Bytes()
}

func (bd *Header) ToArray() []byte {
	bf := new(bytes.Buffer)
	bd.Serialize(bf)
	return bf.Bytes()
}
