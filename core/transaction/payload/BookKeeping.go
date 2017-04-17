package payload

import "io"

type BookKeeping struct {
}

func (a *BookKeeping) Data() []byte {
	return []byte{0}
}

func (a *BookKeeping) Serialize(w io.Writer) error {
	return nil
}

func (a *BookKeeping) Deserialize(r io.Reader) error {
	return nil
}

