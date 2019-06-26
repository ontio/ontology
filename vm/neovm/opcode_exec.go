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

type OpInfo struct {
	Opcode OpCode
	Name   string
}

var (
	OpExecList = [256]OpInfo{
		// control flow
		PUSH0:       {Opcode: PUSH0, Name: "PUSH0"},
		PUSHBYTES1:  {Opcode: PUSHBYTES1, Name: "PUSHBYTES1"},
		PUSHBYTES75: {Opcode: PUSHBYTES75, Name: "PUSHBYTES75"},
		PUSHDATA1:   {Opcode: PUSHDATA1, Name: "PUSHDATA1"},
		PUSHDATA2:   {Opcode: PUSHDATA2, Name: "PUSHDATA2"},
		PUSHDATA4:   {Opcode: PUSHDATA4, Name: "PUSHDATA4"},
		PUSHM1:      {Opcode: PUSHM1, Name: "PUSHM1"},
		PUSH1:       {Opcode: PUSH1, Name: "PUSH1"},
		PUSH2:       {Opcode: PUSH2, Name: "PUSH2"},
		PUSH3:       {Opcode: PUSH3, Name: "PUSH3"},
		PUSH4:       {Opcode: PUSH4, Name: "PUSH4"},
		PUSH5:       {Opcode: PUSH5, Name: "PUSH5"},
		PUSH6:       {Opcode: PUSH6, Name: "PUSH6"},
		PUSH7:       {Opcode: PUSH7, Name: "PUSH7"},
		PUSH8:       {Opcode: PUSH8, Name: "PUSH8"},
		PUSH9:       {Opcode: PUSH9, Name: "PUSH9"},
		PUSH10:      {Opcode: PUSH10, Name: "PUSH10"},
		PUSH11:      {Opcode: PUSH11, Name: "PUSH11"},
		PUSH12:      {Opcode: PUSH12, Name: "PUSH12"},
		PUSH13:      {Opcode: PUSH13, Name: "PUSH13"},
		PUSH14:      {Opcode: PUSH14, Name: "PUSH14"},
		PUSH15:      {Opcode: PUSH15, Name: "PUSH15"},
		PUSH16:      {Opcode: PUSH16, Name: "PUSH16"},

		//Control
		NOP:      {Opcode: NOP, Name: "NOP"},
		JMP:      {Opcode: JMP, Name: "JMP"},
		JMPIF:    {Opcode: JMPIF, Name: "JMPIF"},
		JMPIFNOT: {Opcode: JMPIFNOT, Name: "JMPIFNOT"},
		CALL:     {Opcode: CALL, Name: "CALL"},
		RET:      {Opcode: RET, Name: "RET"},
		APPCALL:  {Opcode: APPCALL, Name: "APPCALL"},
		//TAILCALL: {Opcode: TAILCALL, Name: "TAILCALL"},
		SYSCALL: {Opcode: SYSCALL, Name: "SYSCALL"},
		DCALL:   {Opcode: DCALL, Name: "DCALL"},

		//Stack ops
		DUPFROMALTSTACK: {Opcode: DUPFROMALTSTACK, Name: "DUPFROMALTSTACK"},
		TOALTSTACK:      {Opcode: TOALTSTACK, Name: "TOALTSTACK"},
		FROMALTSTACK:    {Opcode: FROMALTSTACK, Name: "FROMALTSTACK"},
		XDROP:           {Opcode: XDROP, Name: "XDROP"},
		XSWAP:           {Opcode: XSWAP, Name: "XSWAP"},
		XTUCK:           {Opcode: XTUCK, Name: "XTUCK"},
		DEPTH:           {Opcode: DEPTH, Name: "DEPTH"},
		DROP:            {Opcode: DROP, Name: "DROP"},
		DUP:             {Opcode: DUP, Name: "DUP"},
		NIP:             {Opcode: NIP, Name: "NIP"},
		OVER:            {Opcode: OVER, Name: "OVER"},
		PICK:            {Opcode: PICK, Name: "PICK"},
		ROLL:            {Opcode: ROLL, Name: "ROLL"},
		ROT:             {Opcode: ROT, Name: "ROT"},
		SWAP:            {Opcode: SWAP, Name: "SWAP"},
		TUCK:            {Opcode: TUCK, Name: "TUCK"},

		//Splice
		CAT:    {Opcode: CAT, Name: "CAT"},
		SUBSTR: {Opcode: SUBSTR, Name: "SUBSTR"},
		LEFT:   {Opcode: LEFT, Name: "LEFT"},
		RIGHT:  {Opcode: RIGHT, Name: "RIGHT"},
		SIZE:   {Opcode: SIZE, Name: "SIZE"},

		//Bitwiase logic
		INVERT: {Opcode: INVERT, Name: "INVERT"},
		AND:    {Opcode: AND, Name: "AND"},
		OR:     {Opcode: OR, Name: "OR"},
		XOR:    {Opcode: XOR, Name: "XOR"},
		EQUAL:  {Opcode: EQUAL, Name: "EQUAL"},

		//Arithmetic
		INC:         {Opcode: INC, Name: "INC"},
		DEC:         {Opcode: DEC, Name: "DEC"},
		SIGN:        {Opcode: SIGN, Name: "SIGN"},
		NEGATE:      {Opcode: NEGATE, Name: "NEGATE"},
		ABS:         {Opcode: ABS, Name: "ABS"},
		NOT:         {Opcode: NOT, Name: "NOT"},
		NZ:          {Opcode: NZ, Name: "NZ"},
		ADD:         {Opcode: ADD, Name: "ADD"},
		SUB:         {Opcode: SUB, Name: "SUB"},
		MUL:         {Opcode: MUL, Name: "MUL"},
		DIV:         {Opcode: DIV, Name: "DIV"},
		MOD:         {Opcode: MOD, Name: "MOD"},
		SHL:         {Opcode: SHL, Name: "SHL"},
		SHR:         {Opcode: SHR, Name: "SHR"},
		BOOLAND:     {Opcode: BOOLAND, Name: "BOOLAND"},
		BOOLOR:      {Opcode: BOOLOR, Name: "BOOLOR"},
		NUMEQUAL:    {Opcode: NUMEQUAL, Name: "NUMEQUAL"},
		NUMNOTEQUAL: {Opcode: NUMNOTEQUAL, Name: "NUMNOTEQUAL"},
		LT:          {Opcode: LT, Name: "LT"},
		GT:          {Opcode: GT, Name: "GT"},
		LTE:         {Opcode: LTE, Name: "LTE"},
		GTE:         {Opcode: GTE, Name: "GTE"},
		MIN:         {Opcode: MIN, Name: "MIN"},
		MAX:         {Opcode: MAX, Name: "MAX"},
		WITHIN:      {Opcode: WITHIN, Name: "WITHIN"},

		//Crypto
		SHA1:    {Opcode: SHA1, Name: "SHA1"},
		SHA256:  {Opcode: SHA256, Name: "SHA256"},
		HASH160: {Opcode: HASH160, Name: "HASH160"},
		HASH256: {Opcode: HASH256, Name: "HASH256"},
		VERIFY:  {Opcode: VERIFY, Name: "VERIFY"},
		//CHECKSIG:      {Opcode: CHECKSIG, Name: "CHECKSIG"},
		//CHECKMULTISIG: {Opcode: CHECKMULTISIG, Name: "CHECKMULTISIG"},

		//Array
		ARRAYSIZE: {Opcode: ARRAYSIZE, Name: "ARRAYSIZE"},
		PACK:      {Opcode: PACK, Name: "PACK"},
		UNPACK:    {Opcode: UNPACK, Name: "UNPACK"},
		PICKITEM:  {Opcode: PICKITEM, Name: "PICKITEM"},
		SETITEM:   {Opcode: SETITEM, Name: "SETITEM"},
		NEWARRAY:  {Opcode: NEWARRAY, Name: "NEWARRAY"},
		NEWMAP:    {Opcode: NEWMAP, Name: "NEWMAP"},
		NEWSTRUCT: {Opcode: NEWSTRUCT, Name: "NEWSTRUCT"},
		APPEND:    {Opcode: APPEND, Name: "APPEND"},
		REVERSE:   {Opcode: REVERSE, Name: "REVERSE"},
		REMOVE:    {Opcode: REMOVE, Name: "REMOVE"},

		HASKEY: {Opcode: HASKEY, Name: "HASKEY"},
		KEYS:   {Opcode: KEYS, Name: "KEYS"},
		VALUES: {Opcode: VALUES, Name: "VALUES"},

		//Exceptions
		THROW:      {Opcode: THROW, Name: "THROW"},
		THROWIFNOT: {Opcode: THROWIFNOT, Name: "THROWIFNOT"},
	}
)
