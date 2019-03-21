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

type OpCode byte

const (
	// Constants
	PUSH0       OpCode = 0x00 // An empty array of bytes is pushed onto the stack.
	PUSHF       OpCode = PUSH0
	PUSHBYTES1  OpCode = 0x01 // 0x01-0x4B The next opcode bytes is data to be pushed onto the stack
	PUSHBYTES75 OpCode = 0x4B
	PUSHDATA1   OpCode = 0x4C // The next byte contains the number of bytes to be pushed onto the stack.
	PUSHDATA2   OpCode = 0x4D // The next two bytes contain the number of bytes to be pushed onto the stack.
	PUSHDATA4   OpCode = 0x4E // The next four bytes contain the number of bytes to be pushed onto the stack.
	PUSHM1      OpCode = 0x4F // The number -1 is pushed onto the stack.
	PUSH1       OpCode = 0x51 // The number 1 is pushed onto the stack.
	PUSHT       OpCode = PUSH1
	PUSH2       OpCode = 0x52 // The number 2 is pushed onto the stack.
	PUSH3       OpCode = 0x53 // The number 3 is pushed onto the stack.
	PUSH4       OpCode = 0x54 // The number 4 is pushed onto the stack.
	PUSH5       OpCode = 0x55 // The number 5 is pushed onto the stack.
	PUSH6       OpCode = 0x56 // The number 6 is pushed onto the stack.
	PUSH7       OpCode = 0x57 // The number 7 is pushed onto the stack.
	PUSH8       OpCode = 0x58 // The number 8 is pushed onto the stack.
	PUSH9       OpCode = 0x59 // The number 9 is pushed onto the stack.
	PUSH10      OpCode = 0x5A // The number 10 is pushed onto the stack.
	PUSH11      OpCode = 0x5B // The number 11 is pushed onto the stack.
	PUSH12      OpCode = 0x5C // The number 12 is pushed onto the stack.
	PUSH13      OpCode = 0x5D // The number 13 is pushed onto the stack.
	PUSH14      OpCode = 0x5E // The number 14 is pushed onto the stack.
	PUSH15      OpCode = 0x5F // The number 15 is pushed onto the stack.
	PUSH16      OpCode = 0x60 // The number 16 is pushed onto the stack.

	// Flow control
	NOP      OpCode = 0x61 // Does nothing.
	JMP      OpCode = 0x62
	JMPIF    OpCode = 0x63
	JMPIFNOT OpCode = 0x64
	CALL     OpCode = 0x65
	RET      OpCode = 0x66
	APPCALL  OpCode = 0x67
	SYSCALL  OpCode = 0x68
	TAILCALL OpCode = 0x69

	// Stack
	DUPFROMALTSTACK OpCode = 0x6A
	TOALTSTACK      OpCode = 0x6B // Puts the input onto the top of the alt stack. Removes it from the main stack.
	FROMALTSTACK    OpCode = 0x6C // Puts the input onto the top of the main stack. Removes it from the alt stack.
	XDROP           OpCode = 0x6D
	DCALL           OpCode = 0x6E
	XSWAP           OpCode = 0x72
	XTUCK           OpCode = 0x73
	DEPTH           OpCode = 0x74 // Puts the number of stack items onto the stack.
	DROP            OpCode = 0x75 // Removes the top stack item.
	DUP             OpCode = 0x76 // Duplicates the top stack item.
	NIP             OpCode = 0x77 // Removes the second top stack item.
	OVER            OpCode = 0x78 // Copies the second top stack item to the top.
	PICK            OpCode = 0x79 // The item n back in the stack is copied to the top.
	ROLL            OpCode = 0x7A // The item n back in the stack is moved to the top.
	ROT             OpCode = 0x7B // Move third top item on the top of stack.
	SWAP            OpCode = 0x7C // The top two items on the stack are swapped.
	TUCK            OpCode = 0x7D // The item at the top of the stack is copied and inserted before the second-to-top item.

	// Splice
	CAT    OpCode = 0x7E // Concatenates two strings.
	SUBSTR OpCode = 0x7F // Returns a section of a string.
	LEFT   OpCode = 0x80 // Keeps only characters left of the specified point in a string.
	RIGHT  OpCode = 0x81 // Keeps only characters right of the specified point in a string.
	SIZE   OpCode = 0x82 // Returns the length of the input string.

	// Bitwise logic
	INVERT OpCode = 0x83 // Flips all of the bits in the input.
	AND    OpCode = 0x84 // Boolean and between each bit in the inputs.
	OR     OpCode = 0x85 // Boolean or between each bit in the inputs.
	XOR    OpCode = 0x86 // Boolean exclusive or between each bit in the inputs.
	EQUAL  OpCode = 0x87 // Returns 1 if the inputs are exactly equal, 0 otherwise.
	// EQUALVERIFY = 0x88 // Same as EQUAL, but runs VERIFY afterward.
	// RESERVED1 = 0x89 // Transaction is invalid unless occurring in an unexecuted IF branch
	// RESERVED2 = 0x8A // Transaction is invalid unless occurring in an unexecuted IF branch

	// Arithmetic
	// Note: Arithmetic inputs are limited to signed 32-bit integers, but may overflow their output.
	INC         OpCode = 0x8B // 1 is added to the input.
	DEC         OpCode = 0x8C // 1 is subtracted from the input.
	SIGN        OpCode = 0x8D
	NEGATE      OpCode = 0x8F // The sign of the input is flipped.
	ABS         OpCode = 0x90 // The input is made positive.
	NOT         OpCode = 0x91 // If the input is 0 or 1, it is flipped. Otherwise the output will be 0.
	NZ          OpCode = 0x92 // Returns 0 if the input is 0. 1 otherwise.
	ADD         OpCode = 0x93 // a is added to b.
	SUB         OpCode = 0x94 // b is subtracted from a.
	MUL         OpCode = 0x95 // a is multiplied by b.
	DIV         OpCode = 0x96 // a is divided by b.
	MOD         OpCode = 0x97 // Returns the remainder after dividing a by b.
	SHL         OpCode = 0x98 // Shifts a left b bits, preserving sign.
	SHR         OpCode = 0x99 // Shifts a right b bits, preserving sign.
	BOOLAND     OpCode = 0x9A // If both a and b are not 0, the output is 1. Otherwise 0.
	BOOLOR      OpCode = 0x9B // If a or b is not 0, the output is 1. Otherwise 0.
	NUMEQUAL    OpCode = 0x9C // Returns 1 if the numbers are equal, 0 otherwise.
	NUMNOTEQUAL OpCode = 0x9E // Returns 1 if the numbers are not equal, 0 otherwise.
	LT          OpCode = 0x9F // Returns 1 if a is less than b, 0 otherwise.
	GT          OpCode = 0xA0 // Returns 1 if a is greater than b, 0 otherwise.
	LTE         OpCode = 0xA1 // Returns 1 if a is less than or equal to b, 0 otherwise.
	GTE         OpCode = 0xA2 // Returns 1 if a is greater than or equal to b, 0 otherwise.
	MIN         OpCode = 0xA3 // Returns the smaller of a and b.
	MAX         OpCode = 0xA4 // Returns the larger of a and b.
	WITHIN      OpCode = 0xA5 // Returns 1 if x is within the specified range (left-inclusive), 0 otherwise.

	// Crypto
	//RIPEMD160 = 0xA6 // The input is hashed using RIPEMD-160.
	SHA1          OpCode = 0xA7 // The input is hashed using SHA-1.
	SHA256        OpCode = 0xA8 // The input is hashed using SHA-256.
	HASH160       OpCode = 0xA9
	HASH256       OpCode = 0xAA
	CHECKSIG      OpCode = 0xAC // The entire transaction's outputs inputs and script (from the most recently-executed CODESEPARATOR to the end) are hashed. The signature used by CHECKSIG must be a valid signature for this hash and public key. If it is 1 is returned 0 otherwise.
	VERIFY        OpCode = 0xAD
	CHECKMULTISIG OpCode = 0xAE // For each signature and public key pair CHECKSIG is executed. If more public keys than signatures are listed some key/sig pairs can fail. All signatures need to match a public key. If all signatures are valid 1 is returned 0 otherwise. Due to a bug one extra unused value is removed from the stack.

	// Array
	ARRAYSIZE OpCode = 0xC0
	PACK      OpCode = 0xC1
	UNPACK    OpCode = 0xC2
	PICKITEM  OpCode = 0xC3
	SETITEM   OpCode = 0xC4
	NEWARRAY  OpCode = 0xC5
	NEWSTRUCT OpCode = 0xC6
	NEWMAP    OpCode = 0xC7
	APPEND    OpCode = 0xC8
	REVERSE   OpCode = 0xC9
	REMOVE    OpCode = 0xCA
	HASKEY    OpCode = 0xCB
	KEYS      OpCode = 0xCC
	VALUES    OpCode = 0xCD

	//Exception
	THROW      = 0xF0
	THROWIFNOT = 0xF1
)
