package ontid

import (
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"testing"
)

func TestContext(t *testing.T) {
	testcase(t, CaseContext)
}

func CaseContext(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}
	var contexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1"), []byte("https://ontid.ont.io0/did/v1"), []byte("https://ontid.ont.io0/did/v1")}
	context := &Context{
		OntId:    []byte(id),
		Contexts: contexts,
		Index:    1,
		Proof:    []byte{0x01, 0x02},
	}
	sink := common.NewZeroCopySink(nil)
	context.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = addContext(n)
	if err != nil {
		t.Fatal()
	}
	encId, err := encodeID([]byte(id))
	if err != nil {
		t.Fatal()
	}
	key := append(encId, FIELD_CONTEXT)
	res, err := getContexts(n, key)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(common.ToHexString(res[i]))
	}

	contextsJson, err := getContextsWithDefault(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(contextsJson)

	contexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1")}
	context = &Context{
		OntId:    []byte(id),
		Contexts: contexts,
		Index:    1,
		Proof:    []byte{0x01, 0x02},
	}
	sink = common.NewZeroCopySink(nil)
	context.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeContext(n)
	if err != nil {
		t.Fatal()
	}
	res, err = getContexts(n, key)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(common.ToHexString(res[i]))
	}
}
