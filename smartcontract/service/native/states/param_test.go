package states

import (
	"bytes"
	"github.com/ontio/ontology/core/genesis"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestParams_Serialize_Deserialize(t *testing.T) {
	params := new(Params)
	*params = make(map[string]string)
	for i := 0; i < 10; i++ {
		k := "key" + strconv.Itoa(i)
		v := "value" + strconv.Itoa(i)
		(*params)[k] = v
	}
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		t.Fatalf("params serialize error: %v", err)
	}
	deserializeParams := new(Params)
	if err := deserializeParams.Deserialize(bf); err != nil {
		t.Fatalf("params deserialize error: %v", err)
	}
	for i := 0; i < 10; i++ {
		k := "key" + strconv.Itoa(i)
		if (*params)[k] != (*deserializeParams)[k] {
			t.Fatal("params deserialize error")
		}
	}
}

func TestAdmin_Serialize_Deserialize(t *testing.T) {
	admin := new(Admin)
	copy((*admin)[:], genesis.ParamContractAddress[:])
	bf := new(bytes.Buffer)
	if err := admin.Serialize(bf); err != nil {
		t.Fatalf("admin serialize error: %v", err)
	}
	deserializeAdmin := new(Admin)
	if err := deserializeAdmin.Deserialize(bf); err != nil {
		t.Fatal("admin version deserialize error")
	}
	assert.Equal(t, admin, deserializeAdmin)
}
