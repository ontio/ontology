package neovm

import (
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ripemd160"
	"io"
	"testing"
)

func TestHash(t *testing.T) {
	engine := NewExecutionEngine()
	engine.OpCode = HASH160

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	hd160 := Hash(data, engine)

	temp := sha256.Sum256(data)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	assert.Equal(t, hd160, md.Sum(nil))

	temp1 := sha256.Sum256(data)
	data1 := sha256.Sum256(temp1[:])

	engine.OpCode = HASH256
	hd256 := Hash(data, engine)

	assert.Equal(t, data1[:], hd256)
}
