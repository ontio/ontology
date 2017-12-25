package states

import (
	"github.com/Ontology/common"
	"io"
	. "github.com/Ontology/common/serialization"
)

type SpentCoinState struct {
	StateBase
	TransactionHash   common.Uint256
	TransactionHeight uint32
	Items             []*Item
}

type Item struct {
	StateBase
	PrevIndex uint16
	EndHeight uint32
}

func (this *Item) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	err := WriteUint16(w, this.PrevIndex)
	if err != nil {
		return err
	}
	err = WriteUint32(w, this.EndHeight)
	if err != nil {
		return err
	}
	return nil
}

func (this *Item) Deserialize(r io.Reader) error {
	var err error
	err = this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	this.PrevIndex, err = ReadUint16(r)
	if err != nil {
		return err
	}
	this.EndHeight, err = ReadUint32(r)
	if err != nil {
		return err
	}
	return nil
}

func (this *SpentCoinState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	_, err := this.TransactionHash.Serialize(w)
	if err != nil {
		return err
	}
	err = WriteUint32(w, this.TransactionHeight)
	if err != nil {
		return err
	}
	err = WriteUint32(w, uint32(len(this.Items)))
	if err != nil {
		return err
	}
	for _, v := range this.Items {
		err = v.Serialize(w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deserialize is the implement of SignableData interface.
func (this *SpentCoinState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(SpentCoinState)
	}
	var err error
	err = this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	this.TransactionHash.Deserialize(r)
	if err != nil {
		return err
	}
	this.TransactionHeight, err = ReadUint32(r)
	if err != nil {
		return err
	}
	count, err := ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(count); i++ {
		item_ := new(Item)
		err := item_.Deserialize(r)
		if err != nil {
			return err
		}
		this.Items = append(this.Items, item_)
	}
	return nil
}

func (this *SpentCoinState) RemoveItem(i int) {
	this.Items[i] = this.Items[len(this.Items)-1]
	this.Items = this.Items[:len(this.Items)-1]
}

