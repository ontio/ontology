package states

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)


func TestParams_Serialize_Deserialize(t *testing.T) {
	params := new(Params)
	params.Version = 0x2d
	params.ParamList = make([]*Param, 10)
	for i := 0; i < 10; i++ {
		param := &Param{
			Version: 0x2d,
			K:       "key" + strconv.Itoa(i),
			V:       "value" + strconv.Itoa(i),
		}
		params.ParamList[i] = param
	}
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		t.Fatalf("params serialize error: %v", err)
	}
	deserializeParams := new(Params)
	if err := deserializeParams.Deserialize(bf); err != nil {
		t.Fatalf("params deserialize error: %v", err)
	}
	if params.Version != deserializeParams.Version {
		t.Fatal("params version deserialize error")
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("K: %s, V: %s\n", params.ParamList[i].K, deserializeParams.ParamList[i].V)
		if params.ParamList[i].Version != deserializeParams.ParamList[i].Version {
			t.Fatal("params deserialize error")
		}
		if params.ParamList[i].K != deserializeParams.ParamList[i].K {
			t.Fatal("params deserialize error")
		}
		if params.ParamList[i].V != deserializeParams.ParamList[i].V {
			t.Fatal("params deserialize error")
		}
	}
}