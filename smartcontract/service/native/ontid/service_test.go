package ontid

import (
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"testing"
)

func TestService(t *testing.T) {
	testcase(t, CaseService)
}

func CaseService(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}
	service := &ServiceParam{
		OntId:          []byte(id),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;s"),
		Index:          1,
		Proof:          []byte("http;;s;s;s;;s"),
	}

	sink := common.NewZeroCopySink(nil)
	service.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = addService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ := encodeID([]byte(id))
	res, err := getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}
	service = &ServiceParam{
		OntId:          []byte(id),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;ssssss"),
		Index:          1,
		Proof:          []byte("http;;s;s;s;;s"),
	}
	sink = common.NewZeroCopySink(nil)
	service.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = updateService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ = encodeID([]byte(id))
	res, err = getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}

	serviceRemove := &ServiceRemoveParam{
		OntId:     []byte(id),
		ServiceId: []byte("someService"),
		Index:     1,
		Proof:     []byte("http;;s;s;s;;s"),
	}
	sink = common.NewZeroCopySink(nil)
	serviceRemove.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeService(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ = encodeID([]byte(id))
	res, err = getServicesJson(n, encId)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(res[i])
	}

}
