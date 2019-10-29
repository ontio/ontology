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
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/vm/neovm/constants"
	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/vm/neovm/types"
	"golang.org/x/crypto/ripemd160"
)

type VmFeatureFlag struct {
	DisableHasKey  bool // disable haskey, dcall, values opcode
	AllowReaderEOF bool // allow VmReader.ReadBytes got EOF and return 0 bytes
}

func NewExecutor(code []byte, feature VmFeatureFlag) *Executor {
	var engine Executor
	engine.EvalStack = NewValueStack(STACK_LIMIT)
	engine.AltStack = NewValueStack(STACK_LIMIT)
	context := NewExecutionContext(code, feature)
	engine.Context = context
	engine.State = BREAK
	engine.Features = feature
	return &engine
}

type Executor struct {
	EvalStack *ValueStack
	AltStack  *ValueStack
	State     VMState
	Features  VmFeatureFlag
	Callers   []*ExecutionContext
	Context   *ExecutionContext
}

func (self *Executor) PopContext() (*ExecutionContext, error) {
	total := len(self.Callers)
	if total == 0 {
		return nil, errors.ERR_INDEX_OUT_OF_BOUND
	}
	context := self.Callers[total-1]
	self.Callers = self.Callers[:total-1]
	return context, nil
}

func (self *Executor) PushContext(context *ExecutionContext) error {
	if len(self.Callers) >= constants.MAX_INVOCATION_STACK_SIZE {
		return errors.ERR_OVER_STACK_LEN
	}
	self.Callers = append(self.Callers, context)
	return nil
}

func (self *Executor) Execute() error {
	self.State = self.State & (^BREAK)
	for self.Context != nil {
		if self.State == FAULT || self.State == HALT || self.State == BREAK {
			break
		}
		if self.Context == nil {
			break
		}

		opcode, eof := self.Context.ReadOpCode()
		if eof {
			break
		}

		var err error
		self.State, err = self.ExecuteOp(opcode, self.Context)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Executor) checkFeaturesEnabled(opcode OpCode) error {
	switch opcode {
	case HASKEY, KEYS, DCALL, VALUES:
		if self.Features.DisableHasKey {
			return errors.ERR_NOT_SUPPORT_OPCODE
		}
	}

	return nil
}

func (self *Executor) ExecuteOp(opcode OpCode, context *ExecutionContext) (VMState, error) {
	if err := self.checkFeaturesEnabled(opcode); err != nil {
		return FAULT, err
	}

	if opcode >= PUSHBYTES1 && opcode <= PUSHBYTES75 {
		buf, err := context.OpReader.ReadBytes(int(opcode))
		if err != nil {
			return FAULT, err
		}
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
			num, err := context.OpReader.ReadUint16()
			if err != nil {
				return FAULT, err
			}
			numBytes = int(num)
		} else {
			num, err := context.OpReader.ReadUint32()
			if err != nil {
				return FAULT, err
			}
			numBytes = int(num)
		}

		data, err := context.OpReader.ReadBytes(numBytes)
		if err != nil {
			return FAULT, err
		}
		val, err := types.VmValueFromBytes(data)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}
	case PUSHM1, PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16:
		val := int64(opcode) - int64(PUSH1) + 1
		err := self.EvalStack.Push(types.VmValueFromInt64(val))
		if err != nil {
			return FAULT, err
		}
		// Flow control
	case NOP:
		return NONE, nil
	case JMP, JMPIF, JMPIFNOT, CALL:
		if opcode == CALL {
			caller := context.Clone()
			err := caller.SetInstructionPointer(int64(caller.GetInstructionPointer() + 2))
			if err != nil {
				return FAULT, err
			}
			err = self.PushContext(caller)
			if err != nil {
				return FAULT, err
			}
			opcode = JMP
		}

		num, err := context.OpReader.ReadInt16()
		if err != nil {
			return FAULT, err
		}
		offset := int(num)
		offset = context.GetInstructionPointer() + offset - 3

		if offset < 0 || offset > len(context.Code) {
			return FAULT, errors.ERR_FAULT
		}
		var needJmp = true
		if opcode != JMP {
			val, err := self.EvalStack.PopAsBool()
			if err != nil {
				return FAULT, err
			}
			if opcode == JMPIF {
				needJmp = val
			} else {
				needJmp = !val
			}
		}

		if needJmp {
			err := context.SetInstructionPointer(int64(offset))
			if err != nil {
				return FAULT, err
			}
		}
	case DCALL:
		caller := context.Clone()
		err := self.PushContext(caller)
		if err != nil {
			return FAULT, errors.ERR_OVER_STACK_LEN
		}
		target, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		if target < 0 || target >= int64(len(self.Context.Code)) {
			return FAULT, errors.ERR_DCALL_OFFSET_ERROR
		}
		err = self.Context.SetInstructionPointer(target)
		if err != nil {
			return FAULT, err
		}
	case RET:
		// omit handle error is ok, if context stack is empty, self.Context will be nil
		// which will be checked outside before the next opcode call
		self.Context, _ = self.PopContext()
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

	case XDROP: // XDROP is zero based
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
			n = 2
		} else {
			n, err = self.EvalStack.PopAsInt64()
			if err != nil {
				return FAULT, err
			}
		}

		// need clearly define the behave when n == 0 and stack is empty
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
	case INC, DEC, SIGN, NEGATE, ABS:
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
	case NZ:
		x, err := self.EvalStack.PopAsIntValue()
		if err != nil {
			return FAULT, err
		}

		cmp := x.Cmp(types.IntValFromInt(0))
		if cmp == 0 {
			err = self.EvalStack.PushBool(false)
		} else {
			err = self.EvalStack.PushBool(true)
		}

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
	case SHL, SHR:
		x2, err := self.EvalStack.PopAsIntValue()
		if err != nil {
			return FAULT, err
		}
		x1, err := self.EvalStack.PopAsIntValue()
		if err != nil {
			return FAULT, err
		}
		var res types.IntValue
		switch opcode {
		case SHL:
			res, err = x1.Lsh(x2)
			if err != nil {
				return FAULT, err
			}
		case SHR:
			res, err = x1.Rsh(x2)
			if err != nil {
				return FAULT, err
			}
		default:
			panic("unreachable")
		}
		b := types.VmValueFromIntValue(res)
		err = self.EvalStack.Push(b)
		if err != nil {
			return FAULT, err
		}
	case NUMNOTEQUAL, NUMEQUAL:
		// note : pop as bytes to avoid hard-fork because previous version missing check
		// whether the params are a valid 32 byte integer
		left, right, err := self.EvalStack.PopPairAsBytes()
		if err != nil {
			return FAULT, err
		}
		l := common.BigIntFromNeoBytes(left)
		r := common.BigIntFromNeoBytes(right)
		var val bool
		switch opcode {
		case NUMEQUAL:
			val = l.Cmp(r) == 0
		case NUMNOTEQUAL:
			val = l.Cmp(r) != 0
		default:
			panic("unreachable")
		}
		err = self.EvalStack.PushBool(val)
		if err != nil {
			return FAULT, err
		}
	case LT, GT, LTE, GTE:
		leftVal, rightVal, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}
		left, err := leftVal.AsBigInt()
		if err != nil {
			return FAULT, err
		}
		right, err := rightVal.AsBigInt()
		if err != nil {
			return FAULT, err
		}
		var val bool
		switch opcode {
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
		if err != nil {
			return FAULT, err
		}
	// Array
	case ARRAYSIZE:
		val, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}

		var length int64
		if array, err := val.AsArrayValue(); err == nil {
			length = array.Len()
		} else if buf, err := val.AsBytes(); err == nil {
			length = int64(len(buf))
		} else {
			return FAULT, errors.ERR_BAD_TYPE
		}

		err = self.EvalStack.PushInt64(length)
		if err != nil {
			return FAULT, err
		}
	case PACK:
		size, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		if size < 0 {
			return FAULT, errors.ERR_BAD_VALUE
		}
		array := types.NewArrayValue()
		for i := int64(0); i < size; i++ {
			val, err := self.EvalStack.Pop()
			if err != nil {
				return FAULT, err
			}

			err = array.Append(val)
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.Push(types.VmValueFromArrayVal(array))
		if err != nil {
			return FAULT, err
		}
	case UNPACK:
		arr, err := self.EvalStack.PopAsArray()
		if err != nil {
			return FAULT, err
		}
		l := len(arr.Data)
		for i := l - 1; i >= 0; i-- {
			err = self.EvalStack.Push(arr.Data[i])
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.PushInt64(int64(l))
		if err != nil {
			return FAULT, err
		}
	case PICKITEM:
		item, index, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}

		var val types.VmValue
		if array, err := item.AsArrayValue(); err == nil {
			ind, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			if ind < 0 || ind >= array.Len() {
				return FAULT, errors.ERR_INDEX_OUT_OF_BOUND
			}

			val = array.Data[ind]
		} else if struc, err := item.AsStructValue(); err == nil {
			ind, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			if ind < 0 || ind >= struc.Len() {
				return FAULT, errors.ERR_INDEX_OUT_OF_BOUND
			}
			val = struc.Data[ind]
		} else if mapVal, err := item.AsMapValue(); err == nil {
			value, ok, err := mapVal.Get(index)
			if err != nil {
				return FAULT, err
			} else if ok == false {
				// todo: suply a nil value in vm?
				return FAULT, errors.ERR_MAP_NOT_EXIST
			}
			val = value
		} else if buf, err := item.AsBytes(); err == nil {
			ind, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			if ind < 0 || ind >= int64(len(buf)) {
				return FAULT, errors.ERR_INDEX_OUT_OF_BOUND
			}
			val = types.VmValueFromInt64(int64(buf[ind]))
		} else {
			return FAULT, errors.ERR_BAD_TYPE
		}

		err = self.EvalStack.Push(val)
		if err != nil {
			return FAULT, err
		}

	case SETITEM:
		//todo: the original implementation for Struct type may have problem.
		item, index, val, err := self.EvalStack.PopTriple()
		if err != nil {
			return FAULT, err
		}
		if s, err := val.AsStructValue(); err == nil {
			t, err := s.Clone()
			if err != nil {
				return FAULT, err
			}
			val = types.VmValueFromStructVal(t)
		}
		if array, err := item.AsArrayValue(); err == nil {
			ind, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			if ind < 0 || ind >= array.Len() {
				return FAULT, errors.ERR_INDEX_OUT_OF_BOUND
			}

			array.Data[ind] = val
		} else if struc, err := item.AsStructValue(); err == nil {
			ind, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			if ind < 0 || ind >= struc.Len() {
				return FAULT, errors.ERR_INDEX_OUT_OF_BOUND
			}

			struc.Data[ind] = val
		} else if mapVal, err := item.AsMapValue(); err == nil {
			err = mapVal.Set(index, val)
			if err != nil {
				return FAULT, err
			}
		} else {
			return FAULT, errors.ERR_BAD_TYPE
		}
	case NEWARRAY:
		count, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		if count < 0 || count > MAX_ARRAY_SIZE {
			return FAULT, errors.ERR_BAD_VALUE
		}
		array := types.NewArrayValue()
		for i := int64(0); i < count; i++ {
			err = array.Append(types.VmValueFromBool(false))
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.Push(types.VmValueFromArrayVal(array))
		if err != nil {
			return FAULT, err
		}
	case NEWSTRUCT:
		count, err := self.EvalStack.PopAsInt64()
		if err != nil {
			return FAULT, err
		}
		if count < 0 || count > MAX_ARRAY_SIZE {
			return FAULT, errors.ERR_BAD_VALUE
		}
		array := types.NewStructValue()
		for i := int64(0); i < count; i++ {
			err = array.Append(types.VmValueFromBool(false))
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.Push(types.VmValueFromStructVal(array))
		if err != nil {
			return FAULT, err
		}
	case NEWMAP:
		err := self.EvalStack.Push(types.NewMapVmValue())
		if err != nil {
			return FAULT, err
		}
	case APPEND:
		item, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
		if s, err := item.AsStructValue(); err == nil {
			t, err := s.Clone()
			if err != nil {
				return FAULT, err
			}
			item = types.VmValueFromStructVal(t)
		}
		val, err := self.EvalStack.Pop()
		switch val.GetType() {
		case types.StructType:
			array, _ := val.AsStructValue()
			err = array.Append(item)
			if err != nil {
				return FAULT, err
			}
		case types.ArrayType:
			array, _ := val.AsArrayValue()
			err = array.Append(item)
			if err != nil {
				return FAULT, err
			}
		default:
			return FAULT, fmt.Errorf("[executor] ExecuteOp APPEND error, unknown datatype")
		}
	case REVERSE:
		var data []types.VmValue
		item, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
		if array, err := item.AsArrayValue(); err == nil {
			data = array.Data
		} else if struc, err := item.AsStructValue(); err == nil {
			data = struc.Data
		} else {
			return FAULT, errors.ERR_BAD_TYPE
		}

		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
	case REMOVE:
		item, index, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}
		switch item.GetType() {
		case types.MapType:
			value, err := item.AsMapValue()
			if err != nil {
				return FAULT, err
			}
			err = value.Remove(index)
			if err != nil {
				return FAULT, err
			}
		case types.ArrayType:
			value, err := item.AsArrayValue()
			if err != nil {
				return FAULT, err
			}
			i, err := index.AsInt64()
			if err != nil {
				return FAULT, err
			}
			err = value.RemoveAt(i)
			if err != nil {
				return FAULT, err
			}
		default:
			return FAULT, fmt.Errorf("[REMOVE] not support datatype")
		}
	case HASKEY:
		item, key, err := self.EvalStack.PopPair()
		if err != nil {
			return FAULT, err
		}
		mapValue, err := item.AsMapValue()
		if err != nil {
			return FAULT, err
		}
		_, ok, err := mapValue.Get(key)
		if err != nil {
			return FAULT, err
		}
		err = self.EvalStack.Push(types.VmValueFromBool(ok))
		if err != nil {
			return FAULT, err
		}
	case KEYS:
		item, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
		mapValue, err := item.AsMapValue()
		if err != nil {
			return FAULT, err
		}
		keys := mapValue.GetMapSortedKey()
		arr := types.NewArrayValue()
		for _, v := range keys {
			err = arr.Append(v)
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.Push(types.VmValueFromArrayVal(arr))
		if err != nil {
			return FAULT, err
		}
	case VALUES:
		item, err := self.EvalStack.Pop()
		if err != nil {
			return FAULT, err
		}
		mapVal, err := item.AsMapValue()
		if err != nil {
			return FAULT, err
		}
		vals, err := mapVal.GetValues()
		arr := types.NewArrayValue()
		for _, v := range vals {
			err := arr.Append(v)
			if err != nil {
				return FAULT, err
			}
		}
		err = self.EvalStack.Push(types.VmValueFromArrayVal(arr))
		if err != nil {
			return FAULT, err
		}
	case THROW:
		return FAULT, nil
	case THROWIFNOT:
		val, err := self.EvalStack.PopAsBool()
		if err != nil {
			return FAULT, err
		}
		if !val {
			return FAULT, nil
		}
	default:
		return FAULT, errors.ERR_NOT_SUPPORT_OPCODE
	}

	return NONE, nil
}
