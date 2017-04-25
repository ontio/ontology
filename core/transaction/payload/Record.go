package payload

import (
	"DNA/common/serialization"
	. "DNA/errors"
	"errors"
	"io"
)

type Record struct {
	RecordType string
	RecordData []byte
}

func (a *Record) Data() []byte {
	//TODO: implement RegisterRecord.Data()
	return []byte{0}
}

// Serialize is the implement of SignableData interface.
func (a *Record) Serialize(w io.Writer) error {
	err := serialization.WriteVarString(w, a.RecordType)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RecordDetail], RecordType serialize failed.")
	}
	err = serialization.WriteVarBytes(w, a.RecordData)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RecordDetail], RecordData serialize failed.")
	}
	return nil
}

// Deserialize is the implement of SignableData interface.
func (a *Record) Deserialize(r io.Reader) error {
	var err error
	a.RecordType, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(errors.New("[RecordDetail], RecordType deserialize failed."), ErrNoCode, "")
	}
	a.RecordData, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(errors.New("[RecordDetail], RecordData deserialize failed."), ErrNoCode, "")
	}
	return nil
}
