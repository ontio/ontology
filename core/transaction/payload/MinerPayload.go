package payload

import "io"

type MinerPayload struct {
}

func (a *MinerPayload) Data() []byte {
	return []byte{0}
}

func (a *MinerPayload) Serialize(w io.Writer) {
	return
}

func (a *MinerPayload) Deserialize(r io.Reader) error {
	return nil
}

