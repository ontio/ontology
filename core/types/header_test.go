package types

import (
	"bytes"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeader_Serialize(t *testing.T) {
	header := Header{}
	header.Height = 321
	header.Bookkeepers = make([]keypair.PublicKey, 0)
	header.SigData = make([][]byte, 0)
	buf := bytes.NewBuffer(nil)
	err := header.Serialize(buf)
	bs := buf.Bytes()
	assert.Nil(t, err)

	var h2 Header
	h2.Deserialize(buf)
	assert.Equal(t, header, h2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err = h2.Deserialize(buf)

	assert.NotNil(t, err)
}
