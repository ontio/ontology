package ledger

import (
	"errors"
	"io"
)

type Header struct {
	Blockdata *Blockdata
}

//Serialize the blockheader
func (h *Header) Serialize(w io.Writer) {
	h.Blockdata.Serialize(w)
	w.Write([]byte{'0'})

}

func (h *Header) Deserialize(r io.Reader) error {
	h.Blockdata.Deserialize(r)

	var headerFlag [1]byte
	_, err := io.ReadFull(r, headerFlag[:])
	if err != nil {
		return err
	}

	if headerFlag[0] != 0 {
		return errors.New("Format error")
	}

	return nil
}
