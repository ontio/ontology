package oracle

import (
	"bytes"
	"testing"
)

func TestSerialization_Status(t *testing.T) {
	var param1 Status = 10
	bf := new(bytes.Buffer)
	if err := param1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	var param2 Status
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param1 != param2 {
		t.Fatalf("TestSerialization_Status failed")
	}
}

func TestSerialization_RegisterOracleNodeParam(t *testing.T) {
	param1 := &RegisterOracleNodeParam{
		Address:  "aaaa5e502c2c72eb6edaa9516735d518f09c95c3",
		Guaranty: 1000,
	}
	bf := new(bytes.Buffer)
	if err := param1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(RegisterOracleNodeParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param1.Address != param2.Address || param1.Guaranty != param2.Guaranty {
		t.Fatalf("TestSerialization_RegisterOracleNodeParam failed")
	}
}

func TestSerialization_OracleNode(t *testing.T) {
	var status Status = 10
	param1 := &OracleNode{
		Address:  "aaaa5e502c2c72eb6edaa9516735d518f09c95c3",
		Guaranty: 1000,
		Status:   status,
	}
	bf := new(bytes.Buffer)
	if err := param1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(OracleNode)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param1.Address != param2.Address || param1.Guaranty != param2.Guaranty || param1.Status != param2.Status {
		t.Fatalf("TestSerialization_OracleNode failed")
	}
}
