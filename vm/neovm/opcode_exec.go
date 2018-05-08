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

type OpExec struct {
	Opcode    OpCode
	Name      string
	Exec      func(*ExecutionEngine) (VMState, error)
	Validator func(*ExecutionEngine) error
}

var (
	OpExecList = [256]OpExec{
		// control flow
		PUSH0:       {Opcode: PUSH0, Name: "PUSH0", Exec: opPushData},
		PUSHBYTES1:  {Opcode: PUSHBYTES1, Name: "PUSHBYTES1", Exec: opPushData},
		PUSHBYTES75: {Opcode: PUSHBYTES75, Name: "PUSHBYTES75", Exec: opPushData},
		PUSHDATA1:   {Opcode: PUSHDATA1, Name: "PUSHDATA1", Exec: opPushData},
		PUSHDATA2:   {Opcode: PUSHDATA2, Name: "PUSHDATA2", Exec: opPushData},
		PUSHDATA4:   {Opcode: PUSHDATA4, Name: "PUSHDATA4", Exec: opPushData, Validator: validatorPushData4},
		PUSHM1:      {Opcode: PUSHM1, Name: "PUSHM1", Exec: opPushData},
		PUSH1:       {Opcode: PUSH1, Name: "PUSH1", Exec: opPushData},
		PUSH2:       {Opcode: PUSH2, Name: "PUSH2", Exec: opPushData},
		PUSH3:       {Opcode: PUSH3, Name: "PUSH3", Exec: opPushData},
		PUSH4:       {Opcode: PUSH4, Name: "PUSH4", Exec: opPushData},
		PUSH5:       {Opcode: PUSH5, Name: "PUSH5", Exec: opPushData},
		PUSH6:       {Opcode: PUSH6, Name: "PUSH6", Exec: opPushData},
		PUSH7:       {Opcode: PUSH7, Name: "PUSH7", Exec: opPushData},
		PUSH8:       {Opcode: PUSH8, Name: "PUSH8", Exec: opPushData},
		PUSH9:       {Opcode: PUSH9, Name: "PUSH9", Exec: opPushData},
		PUSH10:      {Opcode: PUSH10, Name: "PUSH10", Exec: opPushData},
		PUSH11:      {Opcode: PUSH11, Name: "PUSH11", Exec: opPushData},
		PUSH12:      {Opcode: PUSH12, Name: "PUSH12", Exec: opPushData},
		PUSH13:      {Opcode: PUSH13, Name: "PUSH13", Exec: opPushData},
		PUSH14:      {Opcode: PUSH14, Name: "PUSH14", Exec: opPushData},
		PUSH15:      {Opcode: PUSH15, Name: "PUSH15", Exec: opPushData},
		PUSH16:      {Opcode: PUSH16, Name: "PUSH16", Exec: opPushData},

		//Control
		NOP:      {Opcode: NOP, Name: "NOP", Exec: opNop},
		JMP:      {Opcode: JMP, Name: "JMP", Exec: opJmp},
		JMPIF:    {Opcode: JMPIF, Name: "JMPIF", Exec: opJmp},
		JMPIFNOT: {Opcode: JMPIFNOT, Name: "JMPIFNOT", Exec: opJmp},
		CALL:     {Opcode: CALL, Name: "CALL", Exec: opCall, Validator: validateCall},
		RET:      {Opcode: RET, Name: "RET", Exec: opRet},
		APPCALL:  {Opcode: APPCALL, Name: "APPCALL"},
		//TAILCALL: {Opcode: TAILCALL, Name: "TAILCALL", Exec: opAppCall},
		SYSCALL: {Opcode: SYSCALL, Name: "SYSCALL"},

		//Stack ops
		DUPFROMALTSTACK: {Opcode: DUPFROMALTSTACK, Name: "DUPFROMALTSTACK", Exec: opToDupFromAltStack},
		TOALTSTACK:      {Opcode: TOALTSTACK, Name: "TOALTSTACK", Exec: opToAltStack},
		FROMALTSTACK:    {Opcode: FROMALTSTACK, Name: "FROMALTSTACK", Exec: opFromAltStack},
		XDROP:           {Opcode: XDROP, Name: "XDROP", Exec: opXDrop, Validator: validateXDrop},
		XSWAP:           {Opcode: XSWAP, Name: "XSWAP", Exec: opXSwap, Validator: validateXSwap},
		XTUCK:           {Opcode: XTUCK, Name: "XTUCK", Exec: opXTuck, Validator: validateXTuck},
		DEPTH:           {Opcode: DEPTH, Name: "DEPTH", Exec: opDepth},
		DROP:            {Opcode: DROP, Name: "DROP", Exec: opDrop, Validator: validateCount1},
		DUP:             {Opcode: DUP, Name: "DUP", Exec: opDup},
		NIP:             {Opcode: NIP, Name: "NIP", Exec: opNip, Validator: validateCount2},
		OVER:            {Opcode: OVER, Name: "OVER", Exec: opOver, Validator: validateCount2},
		PICK:            {Opcode: PICK, Name: "PICK", Exec: opPick, Validator: validatePick},
		ROLL:            {Opcode: ROLL, Name: "ROLL", Exec: opRoll, Validator: validateRoll},
		ROT:             {Opcode: ROT, Name: "ROT", Exec: opRot, Validator: validateCount3},
		SWAP:            {Opcode: SWAP, Name: "SWAP", Exec: opSwap, Validator: validateCount2},
		TUCK:            {Opcode: TUCK, Name: "TUCK", Exec: opTuck, Validator: validateCount2},

		//Splice
		CAT:    {Opcode: CAT, Name: "CAT", Exec: opCat, Validator: validateCat},
		SUBSTR: {Opcode: SUBSTR, Name: "SUBSTR", Exec: opSubStr, Validator: validateSubStr},
		LEFT:   {Opcode: LEFT, Name: "LEFT", Exec: opLeft, Validator: validateLeft},
		RIGHT:  {Opcode: RIGHT, Name: "RIGHT", Exec: opRight, Validator: validateRight},
		SIZE:   {Opcode: SIZE, Name: "SIZE", Exec: opSize, Validator: validateCount1},

		//Bitwiase logic
		INVERT: {Opcode: INVERT, Name: "INVERT", Exec: opInvert, Validator: validateCount1},
		AND:    {Opcode: AND, Name: "AND", Exec: opBigIntZip, Validator: validateCount2},
		OR:     {Opcode: OR, Name: "OR", Exec: opBigIntZip, Validator: validateCount2},
		XOR:    {Opcode: XOR, Name: "XOR", Exec: opBigIntZip, Validator: validateCount2},
		EQUAL:  {Opcode: EQUAL, Name: "EQUAL", Exec: opEqual, Validator: validateCount2},

		//Arithmetic
		INC:         {Opcode: INC, Name: "INC", Exec: opBigInt, Validator: validateInc},
		DEC:         {Opcode: DEC, Name: "DEC", Exec: opBigInt, Validator: validateDec},
		SIGN:        {Opcode: SIGN, Name: "SIGN", Exec: opSign, Validator: validateSign},
		NEGATE:      {Opcode: NEGATE, Name: "NEGATE", Exec: opBigInt, Validator: validateCount1},
		ABS:         {Opcode: ABS, Name: "ABS", Exec: opBigInt, Validator: validateCount1},
		NOT:         {Opcode: NOT, Name: "NOT", Exec: opNot, Validator: validateCount1},
		NZ:          {Opcode: NZ, Name: "NZ", Exec: opNz, Validator: validateCount1},
		ADD:         {Opcode: ADD, Name: "ADD", Exec: opBigIntZip, Validator: validateAdd},
		SUB:         {Opcode: SUB, Name: "SUB", Exec: opBigIntZip, Validator: validateSub},
		MUL:         {Opcode: MUL, Name: "MUL", Exec: opBigIntZip, Validator: validateMul},
		DIV:         {Opcode: DIV, Name: "DIV", Exec: opBigIntZip, Validator: validateDiv},
		MOD:         {Opcode: MOD, Name: "MOD", Exec: opBigIntZip, Validator: validateMod},
		SHL:         {Opcode: SHL, Name: "SHL", Exec: opBigIntZip, Validator: validateShiftLeft},
		SHR:         {Opcode: SHR, Name: "SHR", Exec: opBigIntZip, Validator: validateShift},
		BOOLAND:     {Opcode: BOOLAND, Name: "BOOLAND", Exec: opBoolZip, Validator: validateCount2},
		BOOLOR:      {Opcode: BOOLOR, Name: "BOOLOR", Exec: opBoolZip, Validator: validateCount2},
		NUMEQUAL:    {Opcode: NUMEQUAL, Name: "NUMEQUAL", Exec: opBigIntComp, Validator: validateCount2},
		NUMNOTEQUAL: {Opcode: NUMNOTEQUAL, Name: "NUMNOTEQUAL", Exec: opBigIntComp, Validator: validateCount2},
		LT:          {Opcode: LT, Name: "LT", Exec: opBigIntComp, Validator: validateCount2},
		GT:          {Opcode: GT, Name: "GT", Exec: opBigIntComp, Validator: validateCount2},
		LTE:         {Opcode: LTE, Name: "LTE", Exec: opBigIntComp, Validator: validateCount2},
		GTE:         {Opcode: GTE, Name: "GTE", Exec: opBigIntComp, Validator: validateCount2},
		MIN:         {Opcode: MIN, Name: "MIN", Exec: opBigIntZip, Validator: validateCount2},
		MAX:         {Opcode: MAX, Name: "MAX", Exec: opBigIntZip, Validator: validateCount2},
		WITHIN:      {Opcode: WITHIN, Name: "WITHIN", Exec: opWithIn, Validator: validateCount3},

		//Crypto
		SHA1:    {Opcode: SHA1, Name: "SHA1", Exec: opHash, Validator: validateCount1},
		SHA256:  {Opcode: SHA256, Name: "SHA256", Exec: opHash, Validator: validateCount1},
		HASH160: {Opcode: HASH160, Name: "HASH160", Exec: opHash, Validator: validateCount1},
		HASH256: {Opcode: HASH256, Name: "HASH256", Exec: opHash, Validator: validateCount1},
		//CHECKSIG:      {Opcode: CHECKSIG, Name: "CHECKSIG", Exec: opCheckSig, Validator: validateCount2},
		//CHECKMULTISIG: {Opcode: CHECKMULTISIG, Name: "CHECKMULTISIG", Exec: opCheckMultiSig, Validator: validateCount2},

		//Array
		ARRAYSIZE: {Opcode: ARRAYSIZE, Name: "ARRAYSIZE", Exec: opArraySize, Validator: validateCount1},
		PACK:      {Opcode: PACK, Name: "PACK", Exec: opPack, Validator: validatePack},
		UNPACK:    {Opcode: UNPACK, Name: "UNPACK", Exec: opUnpack, Validator: validateUnpack},
		PICKITEM:  {Opcode: PICKITEM, Name: "PICKITEM", Exec: opPickItem, Validator: validatePickItem},
		SETITEM:   {Opcode: SETITEM, Name: "SETITEM", Exec: opSetItem, Validator: validatorSetItem},
		NEWARRAY:  {Opcode: NEWARRAY, Name: "NEWARRAY", Exec: opNewArray, Validator: validateNewArray},
		NEWSTRUCT: {Opcode: NEWSTRUCT, Name: "NEWSTRUCT", Exec: opNewStruct, Validator: validateNewStruct},
		APPEND:    {Opcode: APPEND, Name: "APPEND", Exec: opAppend, Validator: validateAppend},
		REVERSE:   {Opcode: REVERSE, Name: "REVERSE", Exec: opReverse, Validator: validatorReverse},

		//Exceptions
		THROW:      {Opcode: THROW, Name: "THROW", Exec: opThrow},
		THROWIFNOT: {Opcode: THROWIFNOT, Name: "THROWIFNOT", Exec: opThrowIfNot, Validator: validatorThrowIfNot},
	}
)
