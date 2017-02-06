package signature

import (
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
	"GoOnchain/vm"
	"bytes"
	"GoOnchain/crypto"
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

func SignBySigner(data SignableData, signer Signer) []byte {

	return Sign(data,signer.PrivKey(),signer.PubKey().EncodePoint(false)[1:])
}

func GetHashData(data SignableData) []byte {
	//Wjj upd
	b_buf := new(bytes.Buffer)
	//data.SerializeUnsigned(b_buf) //TODO: add SerializeUnsigned method
	return b_buf.Bytes()
}

func GetHashForSigning(data SignableData) []byte {
	//TODO: GetHashForSigning

	return nil
}


func Sign(data SignableData,prikey []byte, pubkey []byte) []byte{

	// FIXME ignore the return error value
	signature, _ := crypto.Sign(prikey, GetHashForSigning(data))
	return signature
}

