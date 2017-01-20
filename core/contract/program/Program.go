package program

import (
	"GoOnchain/common"
	"GoOnchain/common/serialization"
	. "GoOnchain/errors"
	"io"
)

type Program struct {

	//the contract program code,which will be run on VM or specific envrionment
	Code []byte

	//the program code's parameter
	Parameter []byte
}

//Serialize the Program
func (p *Program) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, p.Parameter)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Serialize Code failed.")
	}
	err = serialization.WriteVarBytes(w, p.Code)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Serialize Parameter failed.")
	}
	return nil
}

//Deserialize the Program
func (p *Program) Deserialize(w io.Reader) error {
	val, err := serialization.ReadVarBytes(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Deserialize Code failed.")
	}
	p.Code = val
	p.Parameter, err = serialization.ReadVarBytes(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Deserialize Parameter failed.")
	}
	return nil
}

func (p *Program) CodeHash() common.Uint160 {
	//TODO: implement to code hash
	//new UInt160(script.Sha256().RIPEMD160());

	return common.Uint160{}

}
