package vm

type OpCode byte

type OpItem struct {
	 Opcode   OpCode
	 Len  uint32
	 Data []byte
}

func getOpItem(code []byte,pos uint32) (opitem OpItem,err error){
	//TODO: getOpItem
	return
}

const (
	// Constants
	OP_0   OpCode  = 0x00 // An empty array of bytes is pushed onto the stack. (This is not a no-op: an item is added to the stack.)
	OP_FALSE   OpCode   = OP_0
	OP_PUSHBYTES1  OpCode  = 0x01 // 0x01-0x4B The next  OpCode  bytes is data to be pushed onto the stack
	OP_PUSHBYTES33  OpCode  = 0x21
	OP_PUSHBYTES64  OpCode  = 0x40
	OP_PUSHBYTES75  OpCode  = 0x4B
	OP_PUSHDATA1  OpCode  = 0x4C // The next byte contains the number of bytes to be pushed onto the stack.
	OP_PUSHDATA2  OpCode  = 0x4D // The next two bytes contain the number of bytes to be pushed onto the stack.
	OP_PUSHDATA4  OpCode  = 0x4E // The next four bytes contain the number of bytes to be pushed onto the stack.
	OP_1NEGATE  OpCode  = 0x4F // The number -1 is pushed onto the stack.
	//OP_RESERVED  OpCode  = 0x50 // Transaction is invalid unless occuring in an unexecuted OP_IF branch
	OP_1  OpCode  = 0x51 // The number 1 is pushed onto the stack.
	OP_TRUE  OpCode  = OP_1
	OP_2  OpCode  = 0x52 // The number 2 is pushed onto the stack.
	OP_3  OpCode  = 0x53 // The number 3 is pushed onto the stack.
	OP_4  OpCode  = 0x54 // The number 4 is pushed onto the stack.
	OP_5  OpCode  = 0x55 // The number 5 is pushed onto the stack.
	OP_6  OpCode  = 0x56 // The number 6 is pushed onto the stack.
	OP_7  OpCode  = 0x57 // The number 7 is pushed onto the stack.
	OP_8  OpCode  = 0x58 // The number 8 is pushed onto the stack.
	OP_9  OpCode  = 0x59 // The number 9 is pushed onto the stack.
	OP_10  OpCode  = 0x5A // The number 10 is pushed onto the stack.
	OP_11  OpCode  = 0x5B // The number 11 is pushed onto the stack.
	OP_12  OpCode  = 0x5C // The number 12 is pushed onto the stack.
	OP_13  OpCode  = 0x5D // The number 13 is pushed onto the stack.
	OP_14  OpCode  = 0x5E // The number 14 is pushed onto the stack.
	OP_15  OpCode  = 0x5F // The number 15 is pushed onto the stack.
	OP_16  OpCode  = 0x60 // The number 16 is pushed onto the stack.


	// Flow control
	OP_NOP  OpCode  = 0x61 // Does nothing.
	OP_JMP  OpCode  = 0x62
	OP_JMPIF  OpCode  = 0x63
	OP_JMPIFNOT  OpCode  = 0x64
	OP_CALL  OpCode  = 0x65
	OP_RET  OpCode  = 0x66
	OP_APPCALL  OpCode  = 0x67
	OP_SYSCALL  OpCode  = 0x68
	OP_HALTIFNOT  OpCode  = 0x69
	OP_HALT  OpCode  = 0x6A


	// Stack
	OP_TOALTSTACK  OpCode  = 0x6B // Puts the input onto the top of the alt stack. Removes it from the main stack.
	OP_FROMALTSTACK  OpCode  = 0x6C // Puts the input onto the top of the main stack. Removes it from the alt stack.
	OP_2DROP  OpCode  = 0x6D // Removes the top two stack items.
	OP_2DUP  OpCode  = 0x6E // Duplicates the top two stack items.
	OP_3DUP  OpCode  = 0x6F // Duplicates the top three stack items.
	OP_2OVER  OpCode  = 0x70 // Copies the pair of items two spaces back in the stack to the front.
	OP_2ROT  OpCode  = 0x71 // The fifth and sixth items back are moved to the top of the stack.
	OP_2SWAP  OpCode  = 0x72 // Swaps the top two pairs of items.
	OP_IFDUP  OpCode  = 0x73 // If the top stack value is not 0 duplicate it.
	OP_DEPTH  OpCode  = 0x74 // Puts the number of stack items onto the stack.
	OP_DROP  OpCode  = 0x75 // Removes the top stack item.
	OP_DUP  OpCode  = 0x76 // Duplicates the top stack item.
	OP_NIP  OpCode  = 0x77 // Removes the second-to-top stack item.
	OP_OVER  OpCode  = 0x78 // Copies the second-to-top stack item to the top.
	OP_PICK  OpCode  = 0x79 // The item n back in the stack is copied to the top.
	OP_ROLL  OpCode  = 0x7A // The item n back in the stack is moved to the top.
	OP_ROT  OpCode  = 0x7B // The top three items on the stack are rotated to the left.
	OP_SWAP  OpCode  = 0x7C // The top two items on the stack are swapped.
	OP_TUCK  OpCode  = 0x7D // The item at the top of the stack is copied and inserted before the second-to-top item.


	// Splice
	OP_CAT  OpCode  = 0x7E // Concatenates two strings.
	OP_SUBSTR  OpCode  = 0x7F // Returns a section of a string.
	OP_LEFT  OpCode  = 0x80 // Keeps only characters left of the specified point in a string.
	OP_RIGHT  OpCode  = 0x81 // Keeps only characters right of the specified point in a string.
	OP_SIZE  OpCode  = 0x82 // Returns the length of the input string.


	// Bitwise logic
	OP_INVERT  OpCode  = 0x83 // Flips all of the bits in the input.
	OP_AND  OpCode  = 0x84 // Boolean and between each bit in the inputs.
	OP_OR  OpCode  = 0x85 // Boolean or between each bit in the inputs.
	OP_XOR  OpCode  = 0x86 // Boolean exclusive or between each bit in the inputs.
	OP_EQUAL  OpCode  = 0x87 // Returns 1 if the inputs are exactly equal 0 otherwise.
	//OP_EQUALVERIFY  OpCode  = 0x88 // Same as OP_EQUAL but runs OP_VERIFY afterward.
	//OP_RESERVED1  OpCode  = 0x89 // Transaction is invalid unless occuring in an unexecuted OP_IF branch
	//OP_RESERVED2  OpCode  = 0x8A // Transaction is invalid unless occuring in an unexecuted OP_IF branch

	// Arithmetic
	// Note: Arithmetic inputs are limited to signed 32-bit integers but may overflow their output.
	OP_1ADD  OpCode  = 0x8B // 1 is added to the input.
	OP_1SUB  OpCode  = 0x8C // 1 is subtracted from the input.
	OP_2MUL  OpCode  = 0x8D // The input is multiplied by 2.
	OP_2DIV  OpCode  = 0x8E // The input is divided by 2.
	OP_NEGATE  OpCode  = 0x8F // The sign of the input is flipped.
	OP_ABS  OpCode  = 0x90 // The input is made positive.
	OP_NOT  OpCode  = 0x91 // If the input is 0 or 1 it is flipped. Otherwise the output will be 0.
	OP_0NOTEQUAL  OpCode  = 0x92 // Returns 0 if the input is 0. 1 otherwise.
	OP_ADD  OpCode  = 0x93 // a is added to b.
	OP_SUB  OpCode  = 0x94 // b is subtracted from a.
	OP_MUL  OpCode  = 0x95 // a is multiplied by b.
	OP_DIV  OpCode  = 0x96 // a is divided by b.
	OP_MOD  OpCode  = 0x97 // Returns the remainder after dividing a by b.
	OP_LSHIFT  OpCode  = 0x98 // Shifts a left b bits preserving sign.
	OP_RSHIFT  OpCode  = 0x99 // Shifts a right b bits preserving sign.
	OP_BOOLAND  OpCode  = 0x9A // If both a and b are not 0 the output is 1. Otherwise 0.
	OP_BOOLOR  OpCode  = 0x9B // If a or b is not 0 the output is 1. Otherwise 0.
	OP_NUMEQUAL  OpCode  = 0x9C // Returns 1 if the numbers are equal 0 otherwise.
	//OP_NUMEQUALVERIFY  OpCode  = 0x9D // Same as OP_NUMEQUAL but runs OP_VERIFY afterward.
	OP_NUMNOTEQUAL  OpCode  = 0x9E // Returns 1 if the numbers are not equal 0 otherwise.
	OP_LESSTHAN  OpCode  = 0x9F // Returns 1 if a is less than b 0 otherwise.
	OP_GREATERTHAN  OpCode  = 0xA0 // Returns 1 if a is greater than b 0 otherwise.
	OP_LESSTHANOREQUAL  OpCode  = 0xA1 // Returns 1 if a is less than or equal to b 0 otherwise.
	OP_GREATERTHANOREQUAL  OpCode  = 0xA2 // Returns 1 if a is greater than or equal to b 0 otherwise.
	OP_MIN  OpCode  = 0xA3 // Returns the smaller of a and b.
	OP_MAX  OpCode  = 0xA4 // Returns the larger of a and b.
	OP_WITHIN  OpCode  = 0xA5 // Returns 1 if x is within the specified range (left-inclusive) 0 otherwise.


	// Crypto
	//OP_RIPEMD160  OpCode  = 0xA6 // The input is hashed using RIPEMD-160.
	OP_SHA1  OpCode  = 0xA7 // The input is hashed using SHA-1.
	OP_SHA256  OpCode  = 0xA8 // The input is hashed using SHA-256.
	OP_HASH160  OpCode  = 0xA9
	OP_HASH256  OpCode  = 0xAA
	//OP_CODESEPARATOR  OpCode  = 0xAB // All of the signature checking words will only match signatures to the data after the most recently-executed OP_CODESEPARATOR.
	OP_CHECKSIG  OpCode  = 0xAC // The entire transaction's outputs inputs and script (from the most recently-executed OP_CODESEPARATOR to the end) are hashed. The signature used by OP_CHECKSIG must be a valid signature for this hash and public key. If it is 1 is returned 0 otherwise.
	//OP_CHECKSIGVERIFY  OpCode  = 0xAD // Same as OP_CHECKSIG but OP_VERIFY is executed afterward.
	OP_CHECKMULTISIG  OpCode  = 0xAE // For each signature and public key pair OP_CHECKSIG is executed. If more public keys than signatures are listed some key/sig pairs can fail. All signatures need to match a public key. If all signatures are valid 1 is returned 0 otherwise. Due to a bug one extra unused value is removed from the stack.
	//OP_CHECKMULTISIGVERIFY  OpCode  = 0xAF // Same as OP_CHECKMULTISIG but OP_VERIFY is executed afterward.


	// Array
	OP_ARRAYSIZE  OpCode  = 0xC0
	OP_PACK  OpCode  = 0xC1
	OP_UNPACK  OpCode  = 0xC2
	OP_DISTINCT  OpCode  = 0xC3
	OP_SORT  OpCode  = 0xC4
	OP_REVERSE  OpCode  = 0xC5
	OP_CONCAT  OpCode  = 0xC6
	OP_UNION  OpCode  = 0xC7
	OP_INTERSECT  OpCode  = 0xC8
	OP_EXCEPT  OpCode  = 0xC9
	OP_TAKE  OpCode  = 0xCA
	OP_SKIP  OpCode  = 0xCB
	OP_PICKITEM  OpCode  = 0xCC
	OP_ALL  OpCode  = 0xCD
	OP_ANY  OpCode  = 0xCE
	OP_SUM  OpCode  = 0xCF
	OP_AVERAGE  OpCode  = 0xD0
	OP_MAXITEM  OpCode  = 0xD1
	OP_MINITEM  OpCode  = 0xD2
)


type OpExec struct {
	Opcode   OpCode
	Name  string
	Exec  func(*ExecutionEngine) (VMState,error)
}

var (
	opExecList = [256]OpExec{
		// control flow
		OP_PUSHBYTES1: {OP_PUSHBYTES1, "OP_PUSHBYTES1", opPushData},
		OP_PUSHBYTES33: {OP_PUSHBYTES33, "OP_PUSHBYTES33", opPushData},
		OP_PUSHBYTES64: {OP_PUSHBYTES64, "OP_PUSHBYTES64", opPushData},
		OP_PUSHBYTES75:{OP_PUSHBYTES75, "OP_PUSHBYTES75", opPushData},
		OP_1NEGATE:   	{OP_1NEGATE, "OP_1NEGATE", opPushData},
		OP_0:   	{OP_0, "0", opPushData},
		OP_1:   	{OP_1, "1", opPushData},
		OP_2:   	{OP_2, "2", opPushData},
		OP_3:   	{OP_3, "3", opPushData},
		OP_4:   	{OP_4, "4", opPushData},
		OP_5:   	{OP_5, "5", opPushData},
		OP_6:   	{OP_6, "6", opPushData},
		OP_7:   	{OP_7, "7", opPushData},
		OP_8:   	{OP_8, "8", opPushData},
		OP_9:   	{OP_9, "9", opPushData},
		OP_10:   	{OP_10, "10", opPushData},
		OP_11:   	{OP_11, "11", opPushData},
		OP_12:   	{OP_12, "12", opPushData},
		OP_13:   	{OP_13, "13", opPushData},
		OP_14:   	{OP_14, "14", opPushData},
		OP_15:   	{OP_15, "15", opPushData},
		OP_16:   	{OP_16, "16", opPushData},
		OP_PUSHDATA1: 	{OP_PUSHDATA1, "OP_PUSHDATA1", opPushData},
		OP_PUSHDATA2: 	{OP_PUSHDATA2, "OP_PUSHDATA2", opPushData},
		OP_PUSHDATA4: 	{OP_PUSHDATA4, "OP_PUSHDATA4", opPushData},


		//Control
		OP_NOP: 	{OP_NOP, "OP_NOP", opNop},
		OP_JMP: 	{OP_JMP, "OP_JMP", opJmp},
		OP_JMPIF: 	{OP_JMPIF, "OP_JMPIF", opJmp},
		OP_JMPIFNOT: 	{OP_JMPIFNOT, "OP_JMPIFNOT", opJmp},
		OP_CALL: 	{OP_CALL, "OP_CALL", opCall},
		OP_RET: 	{OP_RET, "OP_RET", opRet},
		//OP_APPCALL: 	{OP_APPCALL, "OP_APPCALL", opAppCall},
		//OP_SYSCALL: 	{OP_SYSCALL, "OP_SYSCALL", opSysCall},
		OP_HALTIFNOT: 	{OP_HALTIFNOT, "OP_HALTIFNOT", opHaltIfNot},
		OP_HALT: 	{OP_HALT, "OP_HALT", opHalts},


		//Stack ops
		OP_TOALTSTACK: 	{OP_TOALTSTACK, "OP_TOALTSTACK", opToAltStack},
		OP_FROMALTSTACK: 	{OP_FROMALTSTACK, "OP_FROMALTSTACK", opFromAltStack},
		OP_2DROP: 		{OP_2DROP, "OP_2DROP", op2Drop},
		OP_2DUP: 		{OP_2DUP, "OP_2DUP", op2Dup},
		OP_3DUP: 		{OP_3DUP, "OP_3DUP", op3Dup},
		OP_2OVER: 		{OP_2OVER, "OP_2OVER", op2Over},
		OP_2ROT: 		{OP_2ROT, "OP_2ROT", op2Rot},
		OP_2SWAP: 		{OP_2SWAP, "OP_2SWAP", op2Swap},
		OP_IFDUP: 		{OP_IFDUP, "OP_IFDUP", opIfDup},
		OP_DEPTH: 		{OP_DEPTH, "OP_DEPTH", opDepth},
		OP_DROP: 		{OP_DROP, "OP_DROP", opDrop},
		OP_DUP: 		{OP_DUP, "OP_DUP", opDup},
		OP_NIP: 		{OP_NIP, "OP_NIP", opNip},
		OP_OVER: 		{OP_OVER, "OP_OVER", opOver},
		OP_PICK: 		{OP_PICK, "OP_PICK", opPick},
		OP_ROLL: 		{OP_ROLL, "OP_ROLL", opRoll},
		OP_ROT: 		{OP_ROT, "OP_ROT", opRot},
		OP_SWAP: 		{OP_SWAP, "OP_SWAP", opSwap},
		OP_TUCK: 		{OP_TUCK, "OP_TUCK", opTuck},
		OP_CAT: 		{OP_CAT, "OP_CAT", opCat},
		OP_SUBSTR: 		{OP_SUBSTR, "OP_SUBSTR", opSubStr},
		OP_LEFT: 		{OP_LEFT, "OP_LEFT", opLeft},
		OP_RIGHT: 		{OP_RIGHT, "OP_RIGHT", opRight},
		OP_SIZE: 		{OP_SIZE, "OP_SIZE", opSize},


		//Bitwiase logic
		OP_INVERT: 		{OP_INVERT, "OP_INVERT", opInvert},
		OP_AND: 		{OP_AND, "OP_AND", opAnd},
		OP_OR: 			{OP_OR, "OP_OR", opOr},
		OP_XOR: 		{OP_XOR, "OP_XOR", opXor},
		OP_EQUAL: 		{OP_EQUAL, "OP_EQUAL", opEqual},


		//Arithmetic
		OP_1ADD: 		{OP_1ADD, "OP_1ADD", op1Add},
		OP_1SUB: 		{OP_1SUB, "OP_1SUB", op1Sub},
		OP_2MUL: 		{OP_2MUL, "OP_2MUL", op2Mul},
		OP_2DIV: 		{OP_2DIV, "OP_2DIV", op2Div},
		OP_NEGATE: 		{OP_NEGATE, "OP_NEGATE", opNegate},
		OP_ABS: 		{OP_ABS, "OP_ABS", opAbs},
		OP_NOT: 		{OP_NOT, "OP_NOT", opNot},
		OP_0NOTEQUAL: 		{OP_0NOTEQUAL, "OP_0NOTEQUAL", op0NotEqual},
		OP_ADD: 		{OP_ADD, "OP_ADD", opAdd},
		OP_SUB: 		{OP_SUB, "OP_SUB", opSub},
		OP_MUL: 		{OP_MUL, "OP_MUL", opMul},
		OP_DIV: 		{OP_DIV, "OP_DIV", opDiv},
		OP_MOD: 		{OP_MOD, "OP_MOD", opMod},
		OP_LSHIFT: 		{OP_LSHIFT, "OP_LSHIFT", opLShift},
		OP_RSHIFT: 		{OP_RSHIFT, "OP_RSHIFT", opRShift},
		OP_BOOLAND: 		{OP_BOOLAND, "OP_BOOLAND", opBoolAnd},
		OP_BOOLOR: 		{OP_BOOLOR, "OP_BOOLOR", opBoolOr},
		OP_NUMEQUAL: 		{OP_NUMEQUAL, "OP_NUMEQUAL", opNumEqual},
		OP_NUMNOTEQUAL: 	{OP_NUMNOTEQUAL, "OP_NUMNOTEQUAL", opNumNotEqual},
		OP_LESSTHAN: 		{OP_LESSTHAN, "OP_LESSTHAN", opLessThan},
		OP_GREATERTHAN:	{OP_GREATERTHAN, "OP_GREATERTHAN", opGreaterThan},
		OP_LESSTHANOREQUAL: 	{OP_LESSTHANOREQUAL, "OP_LESSTHANOREQUAL", opLessThanOrEqual},
		OP_GREATERTHANOREQUAL:{OP_GREATERTHANOREQUAL, "OP_GREATERTHANOREQUAL", opGreaterThanOrEqual},
		OP_MIN:			{OP_MIN, "OP_MIN", opMin},
		OP_MAX:			{OP_MAX, "OP_MAX", opMax},
		OP_WITHIN:		{OP_WITHIN, "OP_WITHIN", opWithIn},


		//Crypto
		OP_SHA1:		{OP_SHA1, "OP_SHA1", opSha1},
		OP_SHA256:		{OP_SHA256, "OP_SHA256", opSha256},
		OP_HASH160:		{OP_HASH160, "OP_HASH160", opHash160},
		OP_HASH256:		{OP_HASH256, "OP_HASH256", opHash256},
		OP_CHECKSIG:		{OP_CHECKSIG, "OP_CHECKSIG", opCheckSig},
		OP_CHECKMULTISIG:	{OP_CHECKMULTISIG, "OP_CHECKMULTISIG", opCheckMultiSig},


		//Array
		OP_ARRAYSIZE:		{OP_ARRAYSIZE, "OP_ARRAYSIZE", opArraySize},
		OP_PACK:		{OP_PACK, "OP_PACK", opPack},
		OP_UNPACK:		{OP_UNPACK, "OP_UNPACK", opUnpack},
		OP_DISTINCT:		{OP_DISTINCT, "OP_DISTINCT", opDistinct},
		OP_SORT:		{OP_SORT, "OP_SORT", opSort},
		OP_REVERSE:		{OP_REVERSE, "OP_REVERSE", opReverse},
		OP_CONCAT:		{OP_CONCAT, "OP_CONCAT", opConcat},
		OP_UNION:		{OP_UNION, "OP_UNION", opUnion},
		OP_INTERSECT:		{OP_INTERSECT, "OP_INTERSECT", opIntersect},
		OP_EXCEPT:		{OP_EXCEPT, "OP_EXCEPT", opExcept},
		OP_TAKE:		{OP_TAKE, "OP_TAKE", opTake},
		OP_SKIP:		{OP_SKIP, "OP_SKIP", opSkip},
		OP_PICKITEM:		{OP_PICKITEM, "OP_PICKITEM", opPickItem},
		OP_ALL:			{OP_ALL, "OP_ALL", opAll},
		OP_ANY:			{OP_ANY, "OP_ANY", opAny},
		OP_SUM:			{OP_SUM, "OP_SUM", opSum},
		OP_AVERAGE:		{OP_AVERAGE, "OP_AVERAGE", opAverage},
		OP_MAXITEM:		{OP_MAXITEM, "OP_MAXITEM", opMaxItem},
		OP_MINITEM:		{OP_MINITEM, "OP_MINITEM", opMinItem},

	}
)

