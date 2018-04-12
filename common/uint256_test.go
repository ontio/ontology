package common

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUint256_Serialize(t *testing.T) {
	var val Uint256
	val[1] = 245
	buf := bytes.NewBuffer(nil)
	err := val.Serialize(buf)
	assert.NotNil(t, err)
}

func TestUint256_Deserialize(t *testing.T) {
	var val Uint256
	val[1] = 245
	buf := bytes.NewBuffer(nil)
	val.Serialize(buf)

	var val2 Uint256
	val2.Deserialize(buf)

	assert.Equal(t, val, val2)

	buf = bytes.NewBuffer([]byte{1, 2, 3})
	err := val2.Deserialize(buf)

	assert.NotNil(t, err)
}

func TestUint256ParseFromBytes(t *testing.T) {
	buf := []byte{1, 2, 3}

	_, err := Uint256ParseFromBytes(buf)

	assert.NotNil(t, err)
}
