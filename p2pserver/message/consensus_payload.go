package message

import (
	"bytes"
	"errors"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookKeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           *crypto.PubKey
	Signature       []byte
	hash            common.Uint256
}

func (cp *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

func (cp *ConsensusPayload) Verify() error {
	buf := new(bytes.Buffer)
	cp.SerializeUnsigned(buf)

	err := crypto.Verify(*cp.Owner, buf.Bytes(), cp.Signature)

	return err
}

func (cp *ConsensusPayload) ToArray() []byte {
	b := new(bytes.Buffer)
	cp.Serialize(b)
	return b.Bytes()
}

func (cp *ConsensusPayload) InventoryType() common.InventoryType {
	return common.CONSENSUS
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	//return sig.GetHashData(cp)
	return []byte{}
}

func (cp *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

func (cp *ConsensusPayload) Serialize(w io.Writer) error {
	err := cp.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	err = cp.Owner.Serialize(w)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, cp.Signature)
	if err != nil {
		return err
	}

	return err
}

func (cp *ConsensusPayload) Deserialize(r io.Reader) error {
	err := cp.DeserializeUnsigned(r)

	pk := new(crypto.PubKey)
	err = pk.DeSerialize(r)
	if err != nil {
		log.Warn("consensus item Owner deserialize failed.")
		return errors.New("consensus item owner deserialize failed. ")
	}
	cp.Owner = pk
	cp.Signature, err = serialization.ReadVarBytes(r)
	return err
}

func (cp *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	serialization.WriteUint32(w, cp.Version)
	cp.PrevHash.Serialize(w)
	serialization.WriteUint32(w, cp.Height)
	serialization.WriteUint16(w, cp.BookKeeperIndex)
	serialization.WriteUint32(w, cp.Timestamp)
	serialization.WriteVarBytes(w, cp.Data)
	return nil
}

func (cp *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	cp.Version, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Version Deserialize failed.")
		return errors.New("consensus item Version Deserialize failed. ")
	}

	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		log.Warn("consensus item preHash Deserialize failed.")
		return errors.New("consensus item preHash Deserialize failed. ")
	}
	cp.PrevHash = *preBlock

	cp.Height, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Height Deserialize failed.")
		return errors.New("consensus item Height Deserialize failed. ")
	}

	cp.BookKeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {
		log.Warn("consensus item BookKeeperIndex Deserialize failed.")
		return errors.New("consensus item BookKeeperIndex Deserialize failed. ")
	}

	cp.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Timestamp Deserialize failed.")
		return errors.New("consensus item Timestamp Deserialize failed. ")
	}

	cp.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("consensus item Data Deserialize failed.")
		return errors.New("consensus item Data Deserialize failed. ")
	}

	return nil
}
