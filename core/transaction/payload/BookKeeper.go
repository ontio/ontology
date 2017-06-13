package payload

import (
	"DNA/common/serialization"
	"DNA/crypto"
	. "DNA/errors"
	"bytes"
	"io"
)

type BookKeeperAction byte

const (
	BookKeeperAction_ADD BookKeeperAction = 0
	BookKeeperAction_SUB BookKeeperAction = 1
)

type BookKeeper struct {
	PubKey *crypto.PubKey
	Action BookKeeperAction
	Cert   []byte
}

func (self *BookKeeper) Data() []byte {
	var buf bytes.Buffer
	self.PubKey.Serialize(&buf)
	buf.WriteByte(byte(self.Action))
	serialization.WriteVarBytes(&buf, self.Cert)

	return buf.Bytes()
}

func (self *BookKeeper) Serialize(w io.Writer) error {
	_, err := w.Write(self.Data())

	return err
}

func (self *BookKeeper) Deserialize(r io.Reader) error {
	self.PubKey = new(crypto.PubKey)
	err := self.PubKey.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], PubKey Deserialize failed.")
	}
	var p [1]byte
	n, err := r.Read(p[:])
	if n == 0 {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Action Deserialize failed.")
	}
	self.Action = BookKeeperAction(p[0])
	self.Cert, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Cert Deserialize failed.")
	}

	return nil
}
