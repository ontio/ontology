package ontid

import (
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"testing"
)

func TestAuthentication(t *testing.T) {
	testcase(t, CaseAuthentication)
}

func CaseAuthentication(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}

	//OntId          []byte
	//IfNewPublicKey bool
	//Index          uint32
	//NewPublicKey   *NewPublicKey
	//SignIndex      uint32
	//Proof          []byte
	newPublicKey := &NewPublicKey{
		key:        nil,
		revoked:    false,
		controller: nil,
	}
	authKeyParam := &AddAuthKeyParam{
		OntId:          []byte(id),
		IfNewPublicKey: false,
		Index:          1,
		NewPublicKey:   newPublicKey,
		SignIndex:      1,
		Proof:          []byte("http;;s;s;s;;s"),
	}
	// 2a6469643a6f6e743a5458625237696f58725a67456571536e696b3843444a3955666d7757505856584a360b736f6d6553657276696365037373730e687474703b3b733b733b733b3b73010e687474703b3b733b733b733b3b73

	sink := common.NewZeroCopySink(nil)
	authKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = addAuthKey(n)
	fmt.Println(common.ToHexString(sink.Bytes()))
	if err != nil {
		t.Fatal()
	}
	encId, _ := encodeID([]byte(id))
	res, err := getAuthentication(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(res)

	//OntId     []byte
	//Index     uint32
	//SignIndex uint32
	//Proof     []byte

	removeAuthKeyParam := &RemoveAuthKeyParam{
		OntId:     []byte(id),
		Index:     1,
		SignIndex: 1,
		Proof:     []byte("http;;s;s;s;;s"),
	}
	sink = common.NewZeroCopySink(nil)
	removeAuthKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeAuthKey(n)
	if err != nil {
		t.Fatal()
	}

	res, err = getAuthentication(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(res)

}
