package common

import (
	"testing"
	"crypto/rand"
)


func BenchmarkZeroCopySource(b *testing.B) {
	const N = 12000
	buf := make([]byte, N)
	rand.Read(buf)

	for i:=0; i< b.N; i++ {
		source := NewZeroCopySource(buf)
		for j := 0; j < N/100; j++  {
			source.NextUint16()
			source.NextByte()
			source.NextUint64()
			source.NextVarUint()
			source.NextBytes(20)
		}
	}

}

