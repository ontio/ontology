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

import (
	"crypto/sha1"
	"crypto/sha256"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/vm/neovm/types"
	"golang.org/x/crypto/ripemd160"
)

func NewExecutor(code []byte) *Executor {
	var engine Executor
	engine.EvalStack = NewValueStack(STACK_LIMIT)
	engine.AltStack = NewValueStack(STACK_LIMIT)
	engine.Context = NewExecutionContext(code)
	engine.State = BREAK
	engine.OpCode = 0
	return &engine
}

type Executor struct {
	EvalStack *ValueStack
	AltStack  *ValueStack
	State     VMState
	//Contexts  []*ExecutionContext
	Context *ExecutionContext
	OpCode  OpCode
	OpExec  OpExec
}

//func (this *Executor) CurrentContext() *ExecutionContext {
//	return this.Contexts[len(this.Contexts)-1]
//}
//
//func (this *Executor) PopContext() (*ExecutionContext, error) {
//	if len(this.Contexts) != 0 {
//		this.Contexts = this.Contexts[:len(this.Contexts)-1]
//	}
//	if len(this.Contexts) != 0 {
//		this.Context = this.CurrentContext()
//	}
//}

//func (this *Executor) PushContext(context *ExecutionContext) {
//	this.Contexts = append(this.Contexts, context)
//	this.Context = this.CurrentContext()
//}

func (self *Executor) Execute() error {
	self.State = self.State & (^BREAK)
	for {
		_, eof := self.ReadOpCode()
		if eof {
			break
		}
		if self.State == FAULT || self.State == HALT || self.State == BREAK {
			break
		}
		err := self.StepInto()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Executor) ReadOpCode() (val OpCode, eof bool) {
	code, err := self.Context.OpReader.ReadByte()
	if err != nil {
		eof = true
		return
	}
	val = OpCode(code)
	self.OpCode = OpCode(code)
	return val, false
}

func (self *Executor) ValidateOp() error {
	opExec := OpExecList[self.OpCode]
	if opExec.Name == "" {
		return errors.ERR_NOT_SUPPORT_OPCODE
	}
	self.OpExec = opExec
	return nil
}

func (self *Executor) StepInto() error {
	state, err := self.ExecuteOp(self.OpCode, self.Context)
	self.State = state
	if err != nil {
		return err
	}
	return nil
}

func (self *Executor) ExecuteOp(opcode OpCode, context *ExecutionContext) (VMState, error) {
	if opcode >= PUSHBYTES1 && opcode <= PUSHBYTES75 {
		buf := context.OpReader.ReadBytes(int(opcode))
		val, err := types.VmValueFromBytes(buf)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
		return NONE, nil
	}

	switch opcode {
	case PUSH0:
		err := self.EvalStack.Push(types.VmValueFromInt64(0))
		if err != nil {
			return FAULT, err
		}
	case PUSHDATA1, PUSHDATA2, PUSHDATA4:
		var numBytes int
		if opcode == PUSHDATA1 {
			d, err := context.OpReader.ReadByte()
			if err != nil {
				return FAULT, err
			}

			numBytes = int(d)
		} else if opcode == PUSHDATA2 {
			numBytes = int(context.OpReader.ReadUint16())
		} else {
			numBytes = int(context.OpReader.ReadInt32())
		}

		data := context.OpReader.ReadBytes(numBytes)
		val, err := types.VmValueFromBytes(data)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case PUSHM1, PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16:
		val := int64(self.OpCode - PUSH1 + 1)
		err := self.EvalStack.Push(types.VmValueFromInt64(val))
		if err != nil {
			return FAULT, err
		}
		// Stack
	case DUPFROMALTSTACK:
		val, err := self.AltStack.Peek(0)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case TOALTSTACK:
		val, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
		err = self.AltStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case FROMALTSTACK:
		val, err := self.AltStack.Pop()
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}

	case XDROP:
		n, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		_, err = self.EvalStack.Remove(n)
		if err != nil {
			return FAULT, err
		}
	case XSWAP:
		n, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Swap(0, n)
		if err != nil {
			return FAULT, err
		}
	case XTUCK:
		n, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}

		val, err := self.EvalStack.Peek(0)
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Insert(n, val)
		if err != nil {
			return FAULT, err
		}
	case DEPTH:
		err := self.EvalStack.PushInt64(int64(self.EvalStack.Count()))
		if err != nil {
			return FAULT, err
		}
	case DROP:
		_, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
	case DUP:
		val, err := self.EvalStack.Peek(0)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case NIP:
		_, val, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case OVER:
		val, err := self.EvalStack.Peek(1)
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case PICK:
		n, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}

		val, err := self.EvalStack.Peek(n)
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case ROLL, ROT:
		var n int64
		var err error
		if opcode == ROT {
			n = 3
		} else {
			n, err = self.EvalStack.PopAsInt64()
			if err != nil {
				return FAULT, err
			}
		}

		// todo: clearly define the behave when n ==0 and stack is empty
		val, err := self.EvalStack.Remove(n)
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case SWAP: // The top two items on the stack are swapped.
		err := self.EvalStack.Swap(0, 1)
		if err != nil {
			return FAULT, err
		}
	case TUCK: // The item at the top of the stack is copied and inserted before the second-to-top item.
		x1, x2, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.PushMany(x2, x1, x2)
		if err != nil {
			return FAULT, err
		}
		// Splice
	case CAT:
		left, right, err := self.EvalStack.PopPairAsBytes()
		if err != nil {
			return FAULT, err
		}

		val := make([]byte, 0, len(left)+len(right))
		val = append(val, left...)
		val = append(val, right...)
		err = self.EvalStack.PushBytes(val)
		if err != nil {
			return FAULT, err
		}
	case SUBSTR:
		start, count, err := self.EvalStack.PopPairAsInt64()
		if err != nil {
			return FAULT, err
		}
		arr, err := self.EvalStack.PopAsBytes()
		if err != nil {
			return FAULT, err
		}

		length := int64(len(arr))
		if start < 0 || start > length {
			return FAULT, errors.ERR_OVER_MAX_ARRAY_SIZE
		}
		if count < 0 || count > length {
			return FAULT, errors.ERR_OVER_MAX_ARRAY_SIZE
		}
		end := start + count
		if end > length {
			return FAULT, errors.ERR_OVER_MAX_ARRAY_SIZE
		}

		b := arr[start:end]
		err = self.EvalStack.PushBytes(b)
		if err != nil {
			return FAULT, err
		}

	case LEFT:
		count, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		arr, err := self.EvalStack.PopAsBytes()
		if err != nil {
			return FAULT, err
		}

		length := int64(len(arr))
		if count < 0 || count > length {
			return FAULT, errors.ERR_OVER_MAX_ARRAY_SIZE
		}

		b := arr[:count]
		err = self.EvalStack.PushBytes(b)
		if err != nil {
			return FAULT, err
		}
	case RIGHT:
		count, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		arr, err := self.EvalStack.PopAsBytes()
		if err != nil {
			return FAULT, err
		}

		length := int64(len(arr))
		if count < 0 || count > length {
			return FAULT, errors.ERR_OVER_MAX_ARRAY_SIZE
		}

		b := arr[length-count:]
		err = self.EvalStack.PushBytes(b)
		if err != nil {
			return FAULT, err
		}
	case SIZE:
		arr, err := self.EvalStack.PopAsBytes()
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.PushInt64(int64(len(arr)))
		if err != nil {
			return FAULT, err
		}
	// Bitwise logic
	case INVERT:
		left, err := self.EvalStack.PopAsIntValue()
		if err != nil {
			return FAULT, err
		}
		val := left.Not()
		err = self.EvalStack.Push(types.VmValueFromIntValue(val))
		if err != nil {
			return FAULT, err
		}
	case AND, OR, XOR:
		left, right, err := self.EvalStack.PopPairAsIntVal()
		if err != nil {
			return FAULT, err
		}

		var val types.IntValue
		switch opcode {
		case AND:
			val, err = left.And(right)
		case OR:
			val, err = left.Or(right)
		case XOR:
			val, err = left.Xor(right)
		default:
			panic("unreachable")
		}
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(types.VmValueFromIntValue(val))
		if err != nil {
			return FAULT, err
		}
	case EQUAL:
		left, right, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.PushBool(left.Equals(right))
		if err != nil {
			return FAULT, err
		}
	case INC, DEC, SIGN, NEGATE, ABS, NZ:
		x, err := self.EvalStack.PopAsIntValue()
		if err != nil {
			return FAULT, err
		}

		var val types.IntValue
		switch opcode {
		case INC:
			val, err = x.Add(types.IntValFromInt(1))
		case DEC:
			val, err = x.Sub(types.IntValFromInt(1))
		case SIGN:
			cmp := x.Cmp(types.IntValFromInt(0))
			val = types.IntValFromInt(int64(cmp))
		case NEGATE:
			val, err = types.IntValFromInt(0).Sub(x)
		case ABS:
			val = x.Abs()
		case NZ:
			cmp := x.Cmp(types.IntValFromInt(0))
			if cmp == 0 {
				val = types.IntValFromInt(0)
			} else {
				val = types.IntValFromInt(1)
			}
		default:
			panic("unreachable")
		}
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.Push(types.VmValueFromIntValue(val))
		if err != nil {
			return FAULT, err
		}
	case ADD, SUB, MUL, DIV, MOD, MAX, MIN:
		left, right, err := self.EvalStack.PopPairAsIntVal()
		if err != nil {
			return FAULT, err
		}
		var val types.IntValue
		switch opcode {
		case ADD:
			val, err = left.Add(right)
		case SUB:
			val, err = left.Sub(right)
		case MUL:
			val, err = left.Mul(right)
		case DIV:
			val, err = left.Div(right)
		case MOD:
			val, err = left.Mod(right)
		case MAX:
			val, err = left.Max(right)
		case MIN:
			val, err = left.Min(right)
		default:
			panic("unreachable")
		}
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(types.VmValueFromIntValue(val))
		if err != nil {
			return FAULT, err
		}
	case NUMEQUAL, NUMNOTEQUAL, LT, GT, LTE, GTE:
		left, right, err := self.EvalStack.PopPairAsIntVal()
		if err != nil {
			return FAULT, err
		}
		var val bool
		switch opcode {
		case NUMEQUAL:
			val = left.Cmp(right) == 0
		case NUMNOTEQUAL:
			val = left.Cmp(right) != 0
		case LT:
			val = left.Cmp(right) < 0
		case GT:
			val = left.Cmp(right) > 0
		case LTE:
			val = left.Cmp(right) <= 0
		case GTE:
			val = left.Cmp(right) >= 0
		default:
			panic("unreachable")
		}
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.PushBool(val)
		if err != nil {
			return FAULT, err
		}

	case BOOLAND, BOOLOR:
		left, right, err := self.EvalStack.PopPairAsBool()
		if err != nil {
			return FAULT, err
		}

		var val bool
		switch opcode {
		case BOOLAND:
			val = left && right
		case BOOLOR:
			val = left || right
		default:
			panic("unreachable")
		}
		err = self.EvalStack.PushBool(val)
		if err != nil {
			return FAULT, err
		}
	case NOT:
		x, err := self.EvalStack.PopAsBool()
		if err != nil {
			return FAULT, err
		}

		err = self.EvalStack.PushBool(!x)
		if err != nil {
			return FAULT, err
		}
	case WITHIN:
		val, left, right, err := self.EvalStack.PopTripleAsIntVal()
		if err != nil {
			return FAULT, err
		}
		v1 := val.Cmp(left)
		v2 := val.Cmp(right)

		err = self.EvalStack.PushBool(v1 >= 0 && v2 < 0)
		if err != nil {
			return FAULT, err
		}
	case SHA1, SHA256, HASH160, HASH256:
		x, err := self.EvalStack.PopAsBytes()
		if err != nil {
			return FAULT, err
		}

		var hash []byte
		switch opcode {
		case SHA1:
			sh := sha1.New()
			sh.Write(x)
			hash = sh.Sum(nil)
		case SHA256:
			sh := sha256.New()
			sh.Write(x)
			hash = sh.Sum(nil)
		case HASH160:
			temp := sha256.Sum256(x)
			md := ripemd160.New()
			md.Write(temp[:])
			hash = md.Sum(nil)
		case HASH256:
			temp := sha256.Sum256(x)
			data := sha256.Sum256(temp[:])
			hash = data[:]
		}
		val, err := types.VmValueFromBytes(hash)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case VERIFY:
		pub, sig, data, err := self.EvalStack.PopTripleAsBytes()
		if err != nil {
			return FAULT, err
		}

		key, err := keypair.DeserializePublicKey(pub)
		if err != nil {
			return FAULT, err
		}

		verErr := signature.Verify(key, data, sig)
		err = self.EvalStack.PushBool(verErr == nil)

	default:
		panic("unimplemented!")
	}

	return NONE, nil
}
