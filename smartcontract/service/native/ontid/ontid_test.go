package ontid

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/testsuite"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

func testcase(t *testing.T, f func(t *testing.T, n *native.NativeService)) {
	testsuite.InvokeNativeContract(t, utils.OntIDContractAddress,
		func(n *native.NativeService) ([]byte, error) {
			f(t, n)
			return nil, nil
		},
	)
}

func TestCaseController(t *testing.T) {
	testcase(t, CaseController)
}

func TestGroupController(t *testing.T) {
	testcase(t, CaseGroupController)
}

func TestRecovery(t *testing.T) {
	testcase(t, CaseRecovery)
}

// Register id with account acc
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

// Register id0 which is controlled by id1
func regControlledID(n *native.NativeService, id0, id1 string, a *account.Account) error {
	// make arguments
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id0))
	sink.WriteVarBytes([]byte(id1))
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a.Address}
	// call
	_, err := regIdWithController(n)
	return err
}

// Test case: register an ID controlled by another ID
func CaseController(t *testing.T, n *native.NativeService) {
	// 1. register the controller
	// create account and id
	a0 := account.NewAccount("")
	id0, err := account.GenerateID()
	assert.Nil(t, err, "generate id0 error")
	assert.Nil(t, regID(n, id0, a0), "register ID error")

	// 2. register the controlled ID
	id1, err := account.GenerateID()
	assert.Nil(t, err, "generate id1 error")
	assert.Nil(t, regControlledID(n, id1, id0, a0), "register by controller error")

	// 3. add attribute
	attr := attribute{
		[]byte("test key"),
		[]byte("test value"),
		[]byte("test type"),
	}
	sink := common.NewZeroCopySink(nil)
	// id
	sink.WriteVarBytes([]byte(id1))
	// attribute
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	_, err = addAttributesByController(n)
	assert.Nil(t, err, "add attribute error")

	// 4. verify signature
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	res, err := verifyController(n)
	assert.Nil(t, err, "verify signature error")
	assert.Equal(t, res, utils.BYTE_TRUE, "verify signature failed")

	// 5. add key
	a1 := account.NewAccount("")
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// key
	pk := keypair.SerializePublicKey(a1.PubKey())
	sink.WriteVarBytes(pk)
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	_, err = addKeyByController(n)
	assert.Nil(t, err, "add key error")

	// 6. remove controller
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// signing key index
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	// set signing address to a1
	n.Tx.SignedAddr = []common.Address{a1.Address}
	_, err = removeController(n)
	assert.Nil(t, err, "remove controller error")
}

func CaseGroupController(t *testing.T, n *native.NativeService) {
	//1. create and register controllers
	id0, err := account.GenerateID()
	assert.Nil(t, err, "create id0 error")
	id1, err := account.GenerateID()
	assert.Nil(t, err, "create id1 error")
	id2, err := account.GenerateID()
	assert.Nil(t, err, "create id2 error")
	a0 := account.NewAccount("")
	a1 := account.NewAccount("")
	a2 := account.NewAccount("")
	assert.Nil(t, regID(n, id0, a0), "register id0 error")
	assert.Nil(t, regID(n, id1, a1), "register id1 error")
	assert.Nil(t, regID(n, id2, a2), "register id2 error")
	// controller group
	g := Group{
		Threshold: 1,
		Members: []interface{}{
			[]byte(id0),
			&Group{
				Threshold: 1,
				Members: []interface{}{
					[]byte(id1),
					[]byte(id2),
				},
			},
		},
	}

	//2. generate and register the controlled id
	id, err := account.GenerateID()
	assert.Nil(t, err, "generate controlled id error")
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id))
	sink.WriteVarBytes(g.Serialize())
	// signers
	signers := []Signer{
		Signer{[]byte(id0), 1},
		Signer{[]byte(id1), 1},
		Signer{[]byte(id2), 1},
	}
	sink.WriteVarBytes(SerializeSigners(signers))
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a0.Address, a1.Address, a2.Address}
	_, err = regIdWithController(n)
	assert.Nil(t, err, "register controlled id error")

	//3. verify signature
	sink.Reset()
	sink.WriteVarBytes([]byte(id))
	signers = []Signer{
		Signer{[]byte(id2), 1},
	}
	sink.WriteVarBytes(SerializeSigners(signers))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	res, err := verifyController(n)
	assert.Nil(t, err, "verify group signature error")
	assert.Equal(t, res, utils.BYTE_TRUE, "verify signature failed")
}

func CaseRecovery(t *testing.T, n *native.NativeService) {
	//1. generate and register id
	id0, err := account.GenerateID()
	assert.Nil(t, err, "generate id0 error")
	a0 := account.NewAccount("")
	assert.Nil(t, regID(n, id0, a0), "register id0 error")
	id1, err := account.GenerateID()
	assert.Nil(t, err, "generate id1 error")
	a1 := account.NewAccount("")
	assert.Nil(t, regID(n, id1, a1), "register id1 error")
	id2, err := account.GenerateID()
	assert.Nil(t, err, "generate id2 error")
	a2 := account.NewAccount("")
	assert.Nil(t, regID(n, id2, a2), "register id2 error")
	//2. set id1 and id2 as id0's recovery id
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id0))
	g := Group{
		Threshold: 1,
		Members:   []interface{}{[]byte(id1), []byte(id2)},
	}
	sink.WriteVarBytes(g.Serialize())
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = addRecovery(n)
	assert.Nil(t, err, "add recovery error")
	//3. add new key by recovery id
	a3 := account.NewAccount("")
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	pk := keypair.SerializePublicKey(a3.PubKey())
	sink.WriteVarBytes(pk)
	s := []Signer{Signer{[]byte(id1), 1}}
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	_, err = addKeyByRecovery(n)
	assert.Nil(t, err, "add key by recovery error")
	//4. remove key
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 1)
	s = []Signer{Signer{[]byte(id2), 1}}
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	_, err = removeKeyByRecovery(n)
	assert.Nil(t, err, "remove key by recovery error")
	//5. verify signature
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a3.Address}
	res, err := verifySignature(n)
	assert.Nil(t, err, "verify signature error")
	assert.Equal(t, utils.BYTE_TRUE, res, "verify signature failed")
	//6. verify signature generated by removed key
	// this should fail
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = verifySignature(n)
	assert.Error(t, err, "signature generated by removed key passed")
}
