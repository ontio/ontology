package auth

import (
	"bytes"
	"testing"
	//"github.com/ontio/ontology/core/genesis"
)

var admin = []byte{0x02, 0x01}
var newAdmin = []byte{0x01, 0x02, 0x03, 0x04}
var funcs = []string{"foo1", "foo2"}
var OntContractAddr = []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

func TestSerialization_Init(t *testing.T) {
	param := &InitContractAdminParam{
		AdminOntID: admin,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(InitContractAdminParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_Transfer(t *testing.T) {
	param := &TransferParam{
		ContractAddr:  OntContractAddr,
		NewAdminOntID: newAdmin,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(TransferParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.NewAdminOntID, param2.NewAdminOntID) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_AssignFuncs(t *testing.T) {
	param := &FuncsToRoleParam{
		ContractAddr: OntContractAddr,
		AdminOntID:   admin,
		Role:         []byte("role"),
		FuncNames:    funcs,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(FuncsToRoleParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_AssignOntIDs(t *testing.T) {
	param := &OntIDsToRoleParam{
		ContractAddr: OntContractAddr,
		AdminOntID:   []byte{0x02, 0x01},
		Role:         []byte("role"),
		Persons:      [][]byte{[]byte{0x03, 0x04, 0x05, 0x06}, []byte{0x07, 0x08, 0x09, 0x0a}},
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(OntIDsToRoleParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_Delegate(t *testing.T) {
	param := &DelegateParam{
		ContractAddr: OntContractAddr,
		From:         []byte{0x03, 0x04, 0x05, 0x06},
		To:           []byte{0x0a, 0x0b, 0x0c, 0x0d},
		Role:         []byte("role"),
		Period:       60 * 60 * 24,
		Level:        3,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(DelegateParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.From, param2.From) != 0 ||
		bytes.Compare(param.To, param2.To) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 ||
		param.Period != param2.Period || param.Level != param2.Level {
		t.Fatalf("failed")
	}
}

func TestSerialization_Withdraw(t *testing.T) {
	param := &WithdrawParam{
		ContractAddr: OntContractAddr,
		Initiator:    []byte{0x03, 0x04, 0x05, 0x06},
		Delegate:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
		Role:         []byte("role"),
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(WithdrawParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.Initiator, param2.Initiator) != 0 ||
		bytes.Compare(param.Delegate, param2.Delegate) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_VerifyToken(t *testing.T) {
	param := &VerifyTokenParam{
		ContractAddr: OntContractAddr,
		Caller:       []byte{0x03, 0x04, 0x05, 0x06},
		Fn:           []byte("foo1"),
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(VerifyTokenParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.ContractAddr, param2.ContractAddr) != 0 ||
		bytes.Compare(param.Caller, param2.Caller) != 0 ||
		bytes.Compare(param.Fn, param2.Fn) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_AuthToken(t *testing.T) {
	param := &AuthToken{
		expireTime: 1000000,
		level:      2,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(AuthToken)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if param.expireTime != param2.expireTime ||
		param.level != param2.level {
		t.Fatalf("failed")
	}
}
