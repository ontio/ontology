package payload

import (
	"DNA/common/serialization"
	"DNA/crypto"
	. "DNA/errors"
	"io"
)

const DataFilePayloadVersion byte = 0x00

type DataFile struct {
	IPFSPath string
	Filename string
	Note     string
	Issuer   *crypto.PubKey
	//TODO: add hash or key to verify data

}

func (a *DataFile) Data(version byte) []byte {
	//TODO: implement RegisterRecord.Data()
	return []byte{0}
}

// Serialize is the implement of SignableData interface.
func (a *DataFile) Serialize(w io.Writer, version byte) error {
	err := serialization.WriteVarString(w, a.IPFSPath)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], IPFSPath serialize failed.")
	}
	err = serialization.WriteVarString(w, a.Filename)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], Filename serialize failed.")
	}
	err = serialization.WriteVarString(w, a.Note)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], Note serialize failed.")
	}
	a.Issuer.Serialize(w)

	return nil
}

// Deserialize is the implement of SignableData interface.
func (a *DataFile) Deserialize(r io.Reader, version byte) error {
	var err error
	a.IPFSPath, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], IPFSPath deserialize failed.")
	}
	a.Filename, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], Filename deserialize failed.")
	}
	a.Note, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], Note deserialize failed.")
	}
	//Issuer     *crypto.PubKey
	a.Issuer = new(crypto.PubKey)
	err = a.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[DataFileDetail], Issuer deserialize failed.")
	}

	return nil
}
