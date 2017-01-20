package signature

import (
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
	"GoOnchain/vm"
	"bytes"
	_ "io"
)

//SignableData describe the data need be signed.
type SignableData interface {
	vm.ISignableObject

	//Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Uint160, error)

	SetPrograms([]*program.Program)

	GetPrograms() []*program.Program

	//TODO: add SerializeUnsigned
	//SerializeUnsigned(io.Writer) error
}

func Sign(data SignableData, signer Signer) ([]byte, error) {

	//TODO: implement Sign()
	return []byte{}, nil
}

func GetHashData(data SignableData) []byte {
	//Wjj upd
	b_buf := new(bytes.Buffer)
	//data.SerializeUnsigned(b_buf) //TODO: add SerializeUnsigned method
	return b_buf.Bytes()
}
