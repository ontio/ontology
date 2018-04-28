package common

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHexAndBytesTransfer(t *testing.T) {
	testBytes := []byte("10, 11, 12, 13, 14, 15, 16, 17, 18, 19")
	stringAfterTrans := ToHexString(testBytes)
	bytesAfterTrans, err := HexToBytes(stringAfterTrans)
	assert.Nil(t, err)
	assert.Equal(t, testBytes, bytesAfterTrans)
}

func TestGetNonce(t *testing.T) {
	nonce1 := GetNonce()
	nonce2 := GetNonce()
	assert.NotEqual(t, nonce1, nonce2)
}

func TestFileExisted(t *testing.T) {
	assert.True(t, FileExisted("common_test.go"))
	assert.True(t, FileExisted("common.go"))
	assert.False(t, FileExisted("../log/log.og"))
	assert.False(t, FileExisted("../log/log.go"))
	assert.True(t, FileExisted("./log/log.go"))
}