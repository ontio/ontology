package payload

import "io"

type IssueAsset struct {

}

func (a *IssueAsset) Data() []byte {
	//TODO: implement IssueAsset.Data()
	return []byte{0}

}

func (a *IssueAsset) Serialize(w io.Writer) error {
	return nil
}

func (a *IssueAsset) Deserialize(r io.Reader) error {
	return nil
}
