package states

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"math/big"
	"testing"
)

func TestTransferFrom_Serialize_Deserialize(t *testing.T) {
	se := "011aac8bb717039d78e9840de35e82ba62723594"
	fr := "011aac8bb717039d78e9840de35e82ba62723594"
	to := "011aac8bb717039d78e9840de35e82ba62723594"

	sb, _ := common.HexToBytes(se)
	fb, _ := common.HexToBytes(fr)
	tb, _ := common.HexToBytes(to)

	sa, _ := common.AddressParseFromBytes(sb)
	fa, _ := common.AddressParseFromBytes(fb)
	ta, _ := common.AddressParseFromBytes(tb)

	tf := &TransferFrom{
		Version: 0,
		Sender:  sa,
		From:    fa,
		To:      ta,
		Value:   big.NewInt(1),
	}

	bf := new(bytes.Buffer)
	if err := tf.Serialize(bf); err != nil {
		t.Fatalf("TransferFrom serialize error: %v", err)
	}

	trf := new(TransferFrom)
	if err := trf.Deserialize(bf); err != nil {
		t.Fatalf("TransferFrom deserialize error: %v", err)
	}
}
