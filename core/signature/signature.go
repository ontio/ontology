package signature

import (
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
	"GoOnchain/crypto"
	"GoOnchain/vm"
	"bytes"
	"crypto/sha256"
	"io"
)

//SignableData describe the data need be signed.
type SignableData interface {
	vm.ISignableObject

	//Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Uint160, error)

	SetPrograms([]*program.Program)

	GetPrograms() []*program.Program

	//TODO: add SerializeUnsigned
	SerializeUnsigned(io.Writer) error
}

func SignBySigner(data SignableData, signer Signer) []byte {

	return Sign(data, signer.PrivKey())
}

func GetHashData(data SignableData) []byte {
	b_buf := new(bytes.Buffer)
	data.SerializeUnsigned(b_buf)
	return b_buf.Bytes()
}

func GetHashForSigning(data SignableData) []byte {
	//TODO: GetHashForSigning
	temp := sha256.Sum256(GetHashData(data))
	return temp[:]
}

func Sign(data SignableData, prikey []byte) []byte {
	// FIXME ignore the return error value
	signature, _ := crypto.Sign(prikey, GetHashForSigning(data))
	return signature
}
