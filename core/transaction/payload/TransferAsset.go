package payload

import "io"

type TransferAsset struct {
}

func (a *TransferAsset) Data() []byte {
	//TODO: implement TransferAsset.Data()
	return []byte{0}

}

func (a *TransferAsset) Serialize(w io.Writer) error {
	return nil
}

func (a *TransferAsset) Deserialize(r io.Reader) error {
	return nil
}
