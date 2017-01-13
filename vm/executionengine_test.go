package vm

import (
	"fmt"
	"testing"
)

func TestExecutionEngine(t *testing.T) {
	var service IApiService
	var crypto ICrypto
	var table IScriptTable
	var signable ISignableObject

	se := NewExecutionEngine( service, crypto, table, signable )
	t.Log( "NewExecutionEngine() test:", se )

	vr := NewVmReader( []byte{0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	//if ( se.Stack.Element[0].array  ){
		t.Errorf("OP_PUSHDATA1 failed. Got ", se.Stack.Element[0].array )
	//}
	t.Log( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[0].array )
	t.Log( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[1].array )
	t.Log( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[2].array )

	vr = NewVmReader( []byte{0x5,0x0,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	se.ExecuteOp( OP_PUSHDATA2, vr )
	fmt.Println( "ExecuteOp() OP_PUSHDATA2 test:", se.Stack.Element[3].array )
}

/*
func ExampleExecutionEngine(){

	var service IApiService
	var crypto ICrypto
	var table IScriptTable
	var signable ISignableObject

	se := NewExecutionEngine( service, crypto, table, signable )
	fmt.Println( "NewExecutionEngine() test:", se )

	vr := NewVmReader( []byte{0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	se.ExecuteOp( OP_PUSHDATA1, vr )
	fmt.Println( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[0].array )
	fmt.Println( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[1].array )
	fmt.Println( "ExecuteOp() OP_PUSHDATA1 test:", se.Stack.Element[2].array )

	vr = NewVmReader( []byte{0x5,0x0,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	se.ExecuteOp( OP_PUSHDATA2, vr )
	fmt.Println( "ExecuteOp() OP_PUSHDATA2 test:", se.Stack.Element[3].array )

	//vr = NewVmReader( []byte{0x5,0x0,0x0,0x0,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	//se.ExecuteOp( OP_PUSHDATA4, vr )
	//fmt.Println( "ExecuteOp() OP_PUSHDATA3 test:", se.Stack.Element[4].array )

	vr = NewVmReader( []byte{0x5,0x0,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	se.ExecuteOp( OP_RET, vr )
	//fmt.Println( "ExecuteOp() OP_RET test:", se.Stack.Element[4].array )

	// output: ok
}
*/

