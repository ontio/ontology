package vm

func opSha1(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	e.Stack.Push(NewStackItem(x.Hash(OP_SHA1,e)))

	return NONE,nil
}

func opSha256(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	e.Stack.Push(NewStackItem(x.Hash(OP_SHA256,e)))

	return NONE,nil
}

func opHash160(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	e.Stack.Push(NewStackItem(x.Hash(OP_HASH160,e)))

	return NONE,nil
}

func opHash256(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	e.Stack.Push(NewStackItem(x.Hash(OP_HASH256,e)))

	return NONE,nil
}

func opCheckSig(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	pubkey := e.Stack.Pop().GetBytes()
	signature := e.Stack.Pop().GetBytes()
	ver,err := e.crypto.VerifySignature(e.signable.GetMessage(),signature,pubkey)
	e.Stack.Push(NewStackItem(ver))

	return NONE,err
}

func opCheckMultiSig(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 4 {return FAULT,nil}
	n :=  int(e.Stack.Pop().ToBigInt().Int64())
	if n < 1 {return FAULT,nil}
	if e.Stack.Count() < n+2 {return FAULT,nil}
	e.nOpCount += n
	if e.nOpCount > MAXSTEPS {return FAULT,nil}

	pubkeys := make([][] byte,n)
	for i := 0; i < n; i++ {pubkeys[i] = e.Stack.Pop().GetBytes()}

	m := int(e.Stack.Pop().ToBigInt().Int64())
	if m < 1 || m > n {return FAULT,nil}
	if e.Stack.Count() < m {return FAULT,nil}

	signatures := make([][] byte,m)
	for i := 0; i < m; i++ {signatures[i] = e.Stack.Pop().GetBytes()}

	message := e.signable.GetMessage()
	fSuccess := true

	for i,j := 0,0; fSuccess && i<m && j<n; {
		ver, err := e.crypto.VerifySignature(message,signatures[i],pubkeys[j])
		if err != nil {
			// TODO: ERROR MSG
		}
		if( ver ) {i++}
		j++
		if m - i > n - j {fSuccess = false}
	}
	e.Stack.Push(NewStackItem(fSuccess))

	return NONE,nil
}
