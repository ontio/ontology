package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArray(t *testing.T) {
	a := NewArrayValue()
	for i := 0; i < 1024; i++ {
		v := VmValueFromInt64(int64(i))
		err := a.Append(v)
		assert.Equal(t, err, nil)
	}
	v := VmValueFromInt64(int64(1024))
	err := a.Append(v)
	assert.NotNil(t, err)
}

func TestArrayValue_RemoveAt(t *testing.T) {
	a := NewArrayValue()
	for i := 0; i < 10; i++ {
		v := VmValueFromInt64(int64(i))
		err := a.Append(v)
		assert.Equal(t, err, nil)
	}
	err := a.RemoveAt(-1)
	assert.NotNil(t, err)
	err = a.RemoveAt(10)
	assert.NotNil(t, err)

	assert.Equal(t, a.Len(), int64(10))
	a.RemoveAt(0)
	assert.Equal(t, a.Len(), int64(9))
}
