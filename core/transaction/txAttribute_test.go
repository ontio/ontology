package transaction

import (
	"bytes"
	"fmt"
	"testing"
)

func TestTxAttribute(t *testing.T) {
	b := new(bytes.Buffer)

	tx := NewTxAttribute(DescriptionUrl, []byte("http:\\www.onchain.com"))
	tx.Serialize(b)
	fmt.Println("Serialize complete")

	tm := TxAttribute{DescriptionUrl, nil, 0}
	tm.Deserialize(b)
	fmt.Println("Deserialize complete.")

	fmt.Printf("Print: Usage= :0x%x,Url Date: %q\n", tm.Usage, tm.Date)
}
