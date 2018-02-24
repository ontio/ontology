package signature

import (
	"bytes"
	"crypto/sha256"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/contract/program"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/vm/neovm/interfaces"
)

//SignableData describe the data need be signed.
type SignableData interface {
	interfaces.ICodeContainer

	////Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Uint160, error)

	SetPrograms([]*program.Program)

	GetPrograms() []*program.Program

	//TODO: add SerializeUnsigned
	SerializeUnsigned(io.Writer) error
}

func SignBySigner(data SignableData, signer Signer) ([]byte, error) {
	log.Debug()
	//fmt.Println("data",data)
	signature, err := crypto.Sign(signer.PrivKey(), GetHashData(data))
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Signature],Sign failed.")
	}
	return signature, nil
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

func Sign(data SignableData, prikey []byte) ([]byte, error) {
	// FIXME ignore the return error value
	signature, err := crypto.Sign(prikey, GetHashData(data))
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Signature],Sign failed.")
	}
	return signature, nil
}
