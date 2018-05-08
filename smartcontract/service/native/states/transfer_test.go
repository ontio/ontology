package states

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"math/big"
	"testing"
)

func TestState_Serialize_Deserialize(t *testing.T) {
	fr := "011aac8bb717039d78e9840de35e82ba62723594"
	to := "011aac8bb717039d78e9840de35e82ba62723595"

	fb, _ := common.HexToBytes(fr)
	tb, _ := common.HexToBytes(to)

	fa, _ := common.AddressParseFromBytes(fb)
	ta, _ := common.AddressParseFromBytes(tb)

	s := &State{
		Version: 0,
		From:    fa,
		To:      ta,
		Value:   big.NewInt(1),
	}

	bf := new(bytes.Buffer)
	if err := s.Serialize(bf); err != nil {
		t.Fatalf("State serialize error: %v", err)
	}

	st := new(State)
	if err := st.Deserialize(bf); err != nil {
		t.Fatalf("State deserialize error: %v", err)
	}
}

func TestTransfers_Serialize_Deserialize(t *testing.T) {
	fr := "011aac8bb717039d78e9840de35e82ba62723594"
	to := "011aac8bb717039d78e9840de35e82ba62723595"

	fb, _ := common.HexToBytes(fr)
	tb, _ := common.HexToBytes(to)

	fa, _ := common.AddressParseFromBytes(fb)
	ta, _ := common.AddressParseFromBytes(tb)

	s := &State{
		Version: 0,
		From:    fa,
		To:      ta,
		Value:   big.NewInt(1),
	}

	ts := &Transfers{
		Version: 0,
		States:  []*State{s},
	}

	bf := new(bytes.Buffer)
	if err := ts.Serialize(bf); err != nil {
		t.Fatalf("Transfers serialize error: %v", err)
	}

	trs := new(Transfers)
	if err := trs.Deserialize(bf); err != nil {
		t.Fatalf("Transfers deserialize error: %v", err)
	}
}
