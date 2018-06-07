package ont

import (
	"bytes"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology/core/types"
)

func TestState_Serialize(t *testing.T) {
	state := State{
		From:  types.AddressFromVmCode([]byte{1, 2, 3}),
		To:    types.AddressFromVmCode([]byte{4, 5, 6}),
		Value: 123,
	}
	bf := new(bytes.Buffer)
	if err := state.Serialize(bf); err != nil {
		t.Fatal("state serialize fail!")
	}

	state2 := State{}
	if err := state2.Deserialize(bf); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, state, state2)
}
