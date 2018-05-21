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
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/vm/neovm/types"
)

func validateCount1(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateCount1]"); err != nil {
		return err
	}
	return nil
}

func validateCount2(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateCount2]"); err != nil {
		return err
	}
	return nil
}

func validateCount3(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 3, "[validateCount3]"); err != nil {
		return err
	}
	return nil
}

func validateDivMod(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateDivMod]"); err != nil {
		return err
	}
	if PeekBigInteger(e).Sign() == 0 {
		return errors.ERR_DIV_MOD_BY_ZERO
	}
	return nil
}

func validateShiftLeft(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateShift]"); err != nil {
		return err
	}

	// x1 << x2
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)

	if x2.Sign() < 0 {
		return errors.ERR_SHIFT_BY_NEG
	}
	if x1.Sign() != 0 && x2.Cmp(big.NewInt(MAX_SIZE_FOR_BIGINTEGER*8)) > 0 {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}

	if CheckBigInteger(new(big.Int).Lsh(x1, uint(x2.Int64()))) == false {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}

	return nil
}

func validateShift(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateShift]"); err != nil {
		return err
	}

	if PeekBigInteger(e).Sign() < 0 {
		return errors.ERR_SHIFT_BY_NEG
	}

	return nil
}

func validatorPushData4(e *ExecutionEngine) error {
	index := e.Context.GetInstructionPointer()
	if index+4 >= len(e.Context.Code) {
		return errors.ERR_OVER_CODE_LEN
	}
	bytesBuffer := bytes.NewBuffer(e.Context.Code[index : index+4])
	var l uint32
	binary.Read(bytesBuffer, binary.LittleEndian, &l)
	if l > MAX_ITEM_SIZE {
		return errors.ERR_OVER_MAX_ITEM_SIZE
	}
	return nil
}

func validateCall(e *ExecutionEngine) error {
	if err := validateInvocationStack(e); err != nil {
		return err
	}
	return nil
}

func validateInvocationStack(e *ExecutionEngine) error {
	if uint32(len(e.Contexts)) >= MAX_INVOCATION_STACK_SIZE {
		return errors.ERR_OVER_STACK_LEN
	}
	return nil
}

func validateOpStack(e *ExecutionEngine, desc string) error {
	total := EvaluationStackCount(e)
	if total < 1 {
		return errors.ERR_UNDER_STACK_LEN
	}
	index := PeekBigInteger(e)
	count := big.NewInt(0)
	if index.Sign() < 0 || count.Add(index, big.NewInt(2)).Cmp(big.NewInt(int64(total))) > 0 {
		return errors.ERR_BAD_VALUE
	}

	return nil
}

func validateXDrop(e *ExecutionEngine) error {
	return validateOpStack(e, "[validateXDrop]")
}

func validateXSwap(e *ExecutionEngine) error {
	return validateOpStack(e, "[validateXSwap]")
}

func validateXTuck(e *ExecutionEngine) error {
	return validateOpStack(e, "[validateXTuck]")
}

func validatePick(e *ExecutionEngine) error {
	return validateOpStack(e, "[validatePick]")
}

func validateRoll(e *ExecutionEngine) error {
	return validateOpStack(e, "[validateRoll]")
}

func validateCat(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateCat]"); err != nil {
		return err
	}
	l := len(PeekNByteArray(0, e)) + len(PeekNByteArray(1, e))
	if uint32(l) > MAX_ITEM_SIZE {
		return errors.ERR_OVER_MAX_ITEM_SIZE
	}
	return nil
}

func validateSubStr(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 3, "[validateSubStr]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	index := PeekNBigInt(1, e)
	if index.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	arr := PeekNByteArray(2, e)
	temp := big.NewInt(0)
	temp.Add(index, count)

	if big.NewInt(int64(len(arr))).Cmp(temp) < 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateLeft(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateLeft]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	arr := PeekNByteArray(1, e)
	if big.NewInt(int64(len(arr))).Cmp(count) < 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateRight(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateRight]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	arr := PeekNByteArray(1, e)
	if big.NewInt(int64(len(arr))).Cmp(count) < 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateInc(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateInc]"); err != nil {
		return err
	}
	x := PeekBigInteger(e)
	if !CheckBigInteger(x) || !CheckBigInteger(new(big.Int).Add(x, big.NewInt(1))) {
		return errors.ERR_BAD_VALUE
	}
	return nil
}

func validateDec(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateDec]"); err != nil {
		return err
	}
	x := PeekBigInteger(e)
	if !CheckBigInteger(x) || (x.Sign() <= 0 && !CheckBigInteger(new(big.Int).Sub(x, big.NewInt(1)))) {
		return errors.ERR_BAD_VALUE
	}
	return nil
}

func validateSign(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateSign]"); err != nil {
		return err
	}
	return nil
}

func validateAdd(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateAdd]"); err != nil {
		return err
	}
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)
	if !CheckBigInteger(x1) || !CheckBigInteger(x2) || !CheckBigInteger(new(big.Int).Add(x1, x2)) {
		return errors.ERR_BAD_VALUE
	}

	return nil
}

func validateSub(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateSub]"); err != nil {
		return err
	}
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)
	if !CheckBigInteger(x1) || !CheckBigInteger(x2) || !CheckBigInteger(new(big.Int).Sub(x1, x2)) {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}
	return nil
}

func validateMul(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateMul]"); err != nil {
		return err
	}
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)
	lx2 := len(types.ConvertBigIntegerToBytes(x2))
	lx1 := len(types.ConvertBigIntegerToBytes(x1))
	if lx2 > MAX_SIZE_FOR_BIGINTEGER || lx1 > MAX_SIZE_FOR_BIGINTEGER || (lx1+lx2) > MAX_SIZE_FOR_BIGINTEGER {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}
	return nil
}

func validateDiv(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateAdd]"); err != nil {
		return err
	}
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)
	if !CheckBigInteger(x2) || !CheckBigInteger(x1) {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}
	if x2.Sign() == 0 {
		return errors.ERR_DIV_MOD_BY_ZERO
	}
	return nil
}

func validateMod(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateMod]"); err != nil {
		return err
	}
	x2 := PeekBigInteger(e)
	x1 := PeekNBigInt(1, e)
	if !CheckBigInteger(x2) || !CheckBigInteger(x1) {
		return errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}
	if x2.Sign() == 0 {
		return errors.ERR_DIV_MOD_BY_ZERO
	}
	return nil
}

func validatePack(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validatePack]"); err != nil {
		return err
	}

	total := EvaluationStackCount(e)
	temp := PeekBigInteger(e)
	count := big.NewInt(0)
	count.Set(temp)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}

	if count.Cmp(big.NewInt(int64(MAX_ARRAY_SIZE))) > 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	count.Add(count, big.NewInt(1))
	if count.Cmp(big.NewInt(int64(total))) > 0 {
		return errors.ERR_OVER_STACK_LEN
	}
	return nil
}

func validateUnpack(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateUnpack]"); err != nil {
		return err
	}
	item := PeekStackItem(e)
	if _, ok := item.(*types.Array); !ok {
		return errors.ERR_NOT_ARRAY
	}
	return nil
}

func validatePickItem(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validatePickItem]"); err != nil {
		return err
	}
	index := PeekBigInteger(e)
	if index.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	item := PeekNStackItem(1, e)
	if item == nil {
		return errors.ERR_BAD_VALUE
	}
	if _, ok := item.(*types.Array); !ok {
		return errors.ERR_NOT_ARRAY
	}
	if index.Cmp(big.NewInt(int64(len(item.GetArray())))) >= 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validatorSetItem(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 3, "[validatorSetItem]"); err != nil {
		return err
	}
	newItem := PeekNStackItem(0, e)
	if newItem == nil {
		return errors.ERR_BAD_VALUE
	}
	index := PeekNBigInt(1, e)
	if index.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	item := PeekNStackItem(2, e)
	if item == nil {
		return errors.ERR_BAD_VALUE
	}
	if index.Cmp(big.NewInt(int64(len(item.GetArray())))) >= 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateNewArray(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateNewArray]"); err != nil {
		return err
	}

	count := PeekBigInteger(e)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	if count.Cmp(big.NewInt(int64(MAX_ARRAY_SIZE))) > 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateNewStruct(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateNewStruct]"); err != nil {
		return err
	}

	count := PeekBigInteger(e)
	if count.Sign() < 0 {
		return errors.ERR_BAD_VALUE
	}
	if count.Cmp(big.NewInt(int64(MAX_ARRAY_SIZE))) > 0 {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	return nil
}

func validateAppend(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateAppend]"); err != nil {
		return err
	}
	arrItem := PeekNStackItem(1, e)
	if _, ok := arrItem.(*types.Array); !ok {
		return errors.ERR_NOT_ARRAY
	}
	return nil
}

func validatorReverse(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validatorReverse]"); err != nil {
		return err
	}
	arrItem := PeekStackItem(e)
	if _, ok := arrItem.(*types.Array); !ok {
		return errors.ERR_NOT_ARRAY
	}
	return nil
}

func validatorThrowIfNot(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validatorThrowIfNot]"); err != nil {
		return err
	}
	return nil
}

func CheckBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(types.ConvertBigIntegerToBytes(value)) > MAX_SIZE_FOR_BIGINTEGER {
		return false
	}
	return true
}

func LogStackTrace(e *ExecutionEngine, needStackCount int, desc string) error {
	stackCount := EvaluationStackCount(e)
	if stackCount < needStackCount {
		return errors.ERR_UNDER_STACK_LEN
	}
	return nil
}
