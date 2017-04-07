package payload

import "io"

type IssueAsset struct {

}

func (a *IssueAsset) Data() []byte {
	//TODO: implement IssueAsset.Data()
	return []byte{0}

}

func (a *IssueAsset) Serialize(w io.Writer) {
	return
}

func (a *IssueAsset) Deserialize(r io.Reader) error {
	return nil
}