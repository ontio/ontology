package program

import (
	"io"
	"GoOnchain/common/serialization"
)
type Program struct {

	//the contract program code,which will be run on VM or specific envrionment
	Code []byte

	//the program code's parameter
	Parameter []byte
}

//Serialize the Program
func (p *Program) Serialize(w io.Writer)  {
	serialization.WriteVarBytes(w,p.Parameter);
	serialization.WriteVarBytes(w,p.Code);
}