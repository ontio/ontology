package common

import (
	"io"
)

//the 64 bit fixed-point number, precise 10^-8
type Fixed64  struct {
	//TODO: implement Fixed64 type
}


func (f *Fixed64) Serialize(w io.Writer) {
	//TODO: implement Fixed64.serialize
}

func (f *Fixed64) GetData() int64 {
	//TODO: implement Fixed64.GetData

	return 0
}