package common

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFixed64_Serialize(t *testing.T) {
	val := Fixed64(10)
	buf := bytes.NewBuffer(nil)
	val.Serialize(buf)
	val2 := Fixed64(0)
	val2.Deserialize(buf)

	assert.Equal(t, val, val2)
}

func TestFixed64_Deserialize(t *testing.T) {
	buf := bytes.NewBuffer([]byte{1, 2, 3})
	val := Fixed64(0)
	err := val.Deserialize(buf)

	assert.NotNil(t, err)

}
