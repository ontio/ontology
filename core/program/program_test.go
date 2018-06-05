package program

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestProgramBuilder_PushBytes(t *testing.T) {
	N := 20000
	builder := ProgramBuilder{}
	for i := 0; i < N; i++ {
		builder.PushNum(uint16(i))
	}
	parser := newProgramParser(builder.Finish())
	for i := 0; i < N; i++ {
		n, err := parser.ReadNum()
		assert.Nil(t, err)
		assert.Equal(t, n, uint16(i))
	}
}

func TestGetProgramInfo(t *testing.T) {
	N := 10
	M := N / 2
	var pubkeys []keypair.PublicKey
	for i := 0; i < N; i++ {
		_, key, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
		pubkeys = append(pubkeys, key)
	}
	list := keypair.NewPublicList(pubkeys)
	sort.Sort(list)
	for i := 0; i < N; i++ {
		pubkeys[i], _ = keypair.DeserializePublicKey(list[i])
	}

	progInfo := ProgramInfo{PubKeys: pubkeys, M: uint16(M)}
	prog, err := ProgramFromMultiPubKey(progInfo.PubKeys, int(progInfo.M))
	assert.Nil(t, err)

	info2, err := GetProgramInfo(prog)
	assert.Nil(t, err)
	assert.Equal(t, progInfo, info2)
}
