package utxo

import (
	"errors"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/errors"
	"io"
)

//define the gas stucture in onchain Ontology
type SpentCoinState struct {
	TransactionHash   common.Uint256
	TransactionHeight uint32
	Items             []*Item
}

// Serialize is the implement of SignableData interface.
func (this *SpentCoinState) Serialize(w io.Writer) error {
	_, err := this.TransactionHash.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[SpentCoinState], TransactionHash serialize failed.")
	}
	err = serialization.WriteUint32(w, this.TransactionHeight)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[SpentCoinState], StartHeight serialize failed.")
	}
	err = serialization.WriteUint32(w, uint32(len(this.Items)))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[SpentCoinState], count serialize failed.")
	}
	for _, v := range this.Items {
		err = v.Serialize(w)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[SpentCoinState], Item serialize failed.")
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
	this.TransactionHash.Deserialize(r)
	if err != nil {
		return NewDetailErr(errors.New("[SpentCoinState], TransactionHash deserialize failed."), ErrNoCode, "")
	}
	this.TransactionHeight, err = serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(errors.New("[SpentCoinState], TransactionHeight deserialize failed."), ErrNoCode, "")
	}
	count, err := serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(errors.New("[SpentCoinState], count deserialize failed."), ErrNoCode, "")
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

type Item struct {
	PrevIndex uint16
	EndHeight uint32
}

// Serialize is the implement of SignableData interface.
func (this *Item) Serialize(w io.Writer) error {
	err := serialization.WriteUint16(w, this.PrevIndex)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Items], PrevIndex serialize failed.")
	}
	err = serialization.WriteUint32(w, this.EndHeight)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Items], EndHeight serialize failed.")
	}
	return nil
}

// Deserialize is the implement of SignableData interface.
func (this *Item) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(Item)
	}
	var err error
	this.PrevIndex, err = serialization.ReadUint16(r)
	if err != nil {
		return NewDetailErr(errors.New("[Items], PrevIndex deserialize failed."), ErrNoCode, "")
	}
	this.EndHeight, err = serialization.ReadUint32(r)
	if err != nil {
		return NewDetailErr(errors.New("[Items], EndHeight deserialize failed."), ErrNoCode, "")
	}
	return nil
}
