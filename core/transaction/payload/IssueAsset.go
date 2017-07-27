package payload

import "io"

const IssueAssetPayloadVersion byte = 0x00

type IssueAsset struct {
}

func (a *IssueAsset) Data(version byte) []byte {
	//TODO: implement IssueAsset.Data()
	return []byte{0}

}

func (a *IssueAsset) Serialize(w io.Writer, version byte) error {
	return nil
}

func (a *IssueAsset) Deserialize(r io.Reader, version byte) error {
	return nil
}
