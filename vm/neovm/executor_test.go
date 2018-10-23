package neovm

import (
	"github.com/ontio/ontology/vm/neovm/types"
	"testing"
)

func BenchmarkNewExecutor(b *testing.B) {
	code := []byte{byte(PUSH1)}

	N := 50000
	for i := 0; i < N; i++ {
		code = append(code, byte(PUSH1), byte(ADD))
	}

	for i := 0; i < b.N; i++ {
		exec := NewExecutor(code)
		err := exec.Execute()
		if err != nil {
			panic(err)
		}
		val, err := exec.EvalStack.PopAsIntValue()
		if err != nil {
			panic(err)
		}
		if val != types.IntValFromInt(int64(N+1)) {
			panic("wrong result")
		}
	}
}

func BenchmarkNative(b *testing.B) {

	N := 50000

	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < N; j++ {
			sum += 1
		}
	}
}
