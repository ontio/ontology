package db

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore("temp.db")
	assert.Nil(t, err)

	_, err = store.GetBestBlock()
	assert.NotNil(t, err)
}

func TestTransactionMeta(t *testing.T) {
	meta := NewTransactionMeta(10, 10)

	for i := uint32(0); i < 10; i++ {
		assert.False(t, meta.IsSpent(i))
		meta.DenoteSpent(i)
	}

	assert.True(t, meta.IsFullSpent())

	for i := uint32(0); i < 10; i++ {
		assert.True(t, meta.IsSpent(i))
		meta.DenoteUnspent(i)
	}
	assert.Equal(t, meta.Height(), uint32(10))

	data := bytes.NewBuffer(nil)
	meta.Serialize(data)
	meta2 := TransactionMeta{}
	meta2.Deserialize(data)
	assert.Equal(t, meta.Height(), meta2.Height())

	for i := uint32(0); i < 10; i++ {
		assert.Equal(t, meta.IsSpent(i), meta2.IsSpent(i))
	}

}
