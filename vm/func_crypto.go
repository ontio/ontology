package vm

import (
	"crypto/sha1"
	"crypto/sha256"
	"hash"
)

func opHash(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	x := AssertStackItem(e.evaluationStack.Pop()).GetByteArray()
	err := pushData(e, Hash(x, e))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opCheckSig(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	pubkey := AssertStackItem(e.evaluationStack.Pop()).GetByteArray()
	signature := AssertStackItem(e.evaluationStack.Pop()).GetByteArray()
	ver, err := e.crypto.VerifySignature(e.scriptContainer.GetMessage(), signature, pubkey)
	err = pushData(e, ver)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opCheckMultiSig(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 4 {
		return FAULT, nil
	}
	n := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if n < 1 {
		return FAULT, nil
	}
	if e.evaluationStack.Count() < n+2 {
		return FAULT, nil
	}
	e.opCount += n
	if e.opCount > e.maxSteps {
		return FAULT, nil
	}

	pubkeys := make([][]byte, n)
	for i := 0; i < n; i++ {
		pubkeys[i] = AssertStackItem(e.evaluationStack.Pop()).GetByteArray()
	}

	m := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if m < 1 || m > n {
		return FAULT, nil
	}
	if e.evaluationStack.Count() < m {
		return FAULT, nil
	}

	signatures := make([][]byte, m)
	for i := 0; i < m; i++ {
		signatures[i] = AssertStackItem(e.evaluationStack.Pop()).GetByteArray()
	}

	message := e.scriptContainer.GetMessage()
	fSuccess := true

	for i, j := 0, 0; fSuccess && i < m && j < n; {
		ver, _ := e.crypto.VerifySignature(message, signatures[i], pubkeys[j])
		if ver {
			i++
		}
		j++
		if m-i > n-j {
			fSuccess = false
		}
	}
	err := pushData(e, fSuccess)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func Hash(b []byte, e *ExecutionEngine) []byte {
	var sh hash.Hash
	var bt []byte
	switch e.opCode {
	case SHA1:
		sh = sha1.New()
		sh.Write(b)
		bt = sh.Sum(nil)
	case SHA256:
		sh = sha256.New()
		sh.Write(b)
		bt = sh.Sum(nil)
	case HASH160:
		bt = e.crypto.Hash160(b)
	case HASH256:
		bt = e.crypto.Hash256(b)
	}
	return bt
}
