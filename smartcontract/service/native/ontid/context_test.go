package ontid

import (
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/testsuite"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"testing"
)

//OntId    []byte
//Contexts [][]byte
//Index    uint32
//Proof    []byte
func testcase(t *testing.T, f func(t *testing.T, n *native.NativeService)) {
	testsuite.InvokeNativeContract(t, utils.OntIDContractAddress,
		func(n *native.NativeService) ([]byte, error) {
			f(t, n)
			return nil, nil
		},
	)
}

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

func regID(n *native.NativeService, id string, a *account.Account) error {
	// make arguments
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id))
	pk := keypair.SerializePublicKey(a.PubKey())
	sink.WriteVarBytes(pk)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a.Address}
	// call
	_, err := regIdWithPublicKey(n)
	return err
}
