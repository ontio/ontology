/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

func opHash(e *ExecutionEngine) (VMState, error) {
	x := PopByteArray(e)
	PushData(e, Hash(x, e))
	return NONE, nil
}

func opCheckSig(e *ExecutionEngine) (VMState, error) {
	pubkey := PopByteArray(e)
	signature := PopByteArray(e)

	// TODO use Hash for VerifySignature data
	panic("need reimplement sig data should be hash")
	ver, err := e.crypto.VerifySignature(e.codeContainer.GetMessage(), signature, pubkey)
	if err != nil {
		return FAULT, err
	}
	PushData(e, ver)
	return NONE, nil
}

func opCheckMultiSig(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	if n < 1 {
		return FAULT, nil
	}
	if Count(e) < n+2 {
		return FAULT, nil
	}
	e.opCount += n

	pubkeys := make([][]byte, n)
	for i := 0; i < n; i++ {
		pubkeys[i] = PopByteArray(e)
	}

	m := PopInt(e)
	if m < 1 || m > n {
		return FAULT, nil
	}

	signatures := make([][]byte, m)
	for i := 0; i < m; i++ {
		signatures[i] = PopByteArray(e)
	}

	message := e.codeContainer.GetMessage()
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
	PushData(e, fSuccess)
	return NONE, nil
}
