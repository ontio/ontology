package transaction

import (
	"io"
)

//TxAttribute descirbte the specific attributes of transcation
type TxAttribute struct {
	//TODO: defaine TxAttribute type

}



func (u *TxAttribute) Serialize(w io.Writer)  {
	//TODO: implement TxAttribute.Serialize()

}

func (tx *TxAttribute) Deserialize(r io.Reader) error  {
	//TODO；TxAttribute Deserialize

	return nil
}
