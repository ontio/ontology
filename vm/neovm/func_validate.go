package neovm

import (
	"bytes"
	"encoding/binary"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/vm/neovm/types"
	"math/big"

	"fmt"
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
		return ErrDivModByZero
	}
	return nil
}

func validateShift(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateShift]"); err != nil {
		return err
	}

	if PeekBigInteger(e).Sign() < 0 {
		return ErrShiftByNeg
	}

	return nil
}

func validatorPushData4(e *ExecutionEngine) error {
	index := e.context.GetInstructionPointer()
	if index+4 >= len(e.context.Code) {
		return ErrOverCodeLen
	}
	bytesBuffer := bytes.NewBuffer(e.context.Code[index : index+4])
	var l uint32
	binary.Read(bytesBuffer, binary.LittleEndian, &l)
	if l > MaxItemSize {
		return ErrOverMaxItemSize
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
	if uint32(e.invocationStack.Count()) >= MaxInvovationStackSize {
		return ErrOverStackLen
	}
	return nil
}

func validateAppCall(e *ExecutionEngine) error {
	if err := validateInvocationStack(e); err != nil {
		return err
	}
	if e.table == nil {
		return ErrTableIsNil
	}
	return nil
}

func validateSysCall(e *ExecutionEngine) error {
	if e.service == nil {
		return ErrServiceIsNil
	}
	return nil
}

func validateOpStack(e *ExecutionEngine, desc string) error {
	total := EvaluationStackCount(e)
	if total < 1 {
		log.Error(desc, total, 1)
		return ErrUnderStackLen
	}
	index := PeekBigInteger(e)
	count := big.NewInt(0)
	if index.Sign() < 0 || count.Add(index, big.NewInt(2)).Cmp(big.NewInt(int64(total))) > 0 {
		log.Error(desc, " index < 0 || index > EvaluationStackCount(e)-2")
		return ErrBadValue
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
	if uint32(l) > MaxItemSize {
		log.Error("[validateCat] uint32(l) > MaxItemSize")
		return ErrOverMaxItemSize
	}
	return nil
}

func validateSubStr(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 3, "[validateSubStr]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		log.Error("[validateSubStr] count < 0")
		return ErrBadValue
	}
	index := PeekNBigInt(1, e)
	if index.Sign() < 0 {
		log.Error("[validateSubStr] index < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(2, e)
	temp := big.NewInt(0)
	temp.Add(index, count)

	if big.NewInt(int64(len(arr))).Cmp(temp) < 0 {
		log.Error("[validateSubStr] len(arr) < index + count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validateLeft(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateLeft]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		log.Error("[validateLeft] count < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(1, e)
	if big.NewInt(int64(len(arr))).Cmp(count) < 0 {
		log.Error("[validateLeft] len(arr) < count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validateRight(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validateRight]"); err != nil {
		return err
	}
	count := PeekNBigInt(0, e)
	if count.Sign() < 0 {
		log.Error("[validateRight] count < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(1, e)
	if big.NewInt(int64(len(arr))).Cmp(count) < 0 {
		log.Error("[validateRight] len(arr) < count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validateInc(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateInc]"); err != nil {
		return err
	}
	x := PeekBigInteger(e)
	if !CheckBigInteger(x) || !CheckBigInteger(new(big.Int).Add(x, big.NewInt(1))) {
		log.Error("[validateInc] CheckBigInteger fail")
		return ErrBadValue
	}
	return nil
}

func validateDec(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateDec]"); err != nil {
		return err
	}
	x := PeekBigInteger(e)
	if !CheckBigInteger(x) || (x.Sign() <= 0 && !CheckBigInteger(new(big.Int).Sub(x, big.NewInt(1)))) {
		log.Error("[validateDec] CheckBigInteger fail")
		return ErrBadValue
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
		log.Error("[validateAdd] CheckBigInteger fail")
		return ErrBadValue
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
		log.Error("[validateAdd] CheckBigInteger fail")
		return ErrOverMaxBigIntegerSize
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
	if lx2 > MaxSizeForBigInteger || lx1 > MaxSizeForBigInteger || (lx1+lx2) > MaxSizeForBigInteger {
		log.Error("[validateMul] CheckBigInteger fail")
		return ErrOverMaxBigIntegerSize
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
		log.Error("[validateDiv] CheckBigInteger fail")
		return ErrOverMaxBigIntegerSize
	}
	if x2.Sign() == 0 {
		return ErrDivModByZero
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
		log.Error("[validateMod] CheckBigInteger fail")
		return ErrOverMaxBigIntegerSize
	}
	if x2.Sign() == 0 {
		return ErrDivModByZero
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
		return ErrBadValue
	}

	if count.Cmp(big.NewInt(int64(MaxArraySize))) > 0 {
		log.Error("[validateRight] uint32(count) > MaxArraySize")
		return ErrOverMaxArraySize
	}
	count.Add(count, big.NewInt(1))
	if count.Cmp(big.NewInt(int64(total))) > 0 {
		log.Error("[validateRight] count+2 > EvaluationStackCount(e)")
		return ErrOverStackLen
	}
	return nil
}

func validatePickItem(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 2, "[validatePickItem]"); err != nil {
		return err
	}
	index := PeekBigInteger(e)
	if index.Sign() < 0 {
		log.Error("[validatePickItem] index < 0")
		return ErrBadValue
	}
	item := PeekN(1, e)
	if item == nil {
		log.Error("[validatePickItem] item = nil")
		return ErrBadValue
	}
	stackItem := item.GetStackItem()
	if _, ok := stackItem.(*types.Array); !ok {
		log.Error("[validatePickItem] ErrNotArray")
		return ErrNotArray
	}
	if index.Cmp(big.NewInt(int64(len(stackItem.GetArray())))) >= 0 {
		log.Error("[validatePickItem] index >= len(stackItem.GetArray())")
		return ErrOverMaxArraySize
	}
	return nil
}

func validatorSetItem(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 3, "[validatorSetItem]"); err != nil {
		return err
	}
	newItem := PeekN(0, e)
	if newItem == nil {
		log.Error("[validatorSetItem] newItem = nil")
		return ErrBadValue
	}
	index := PeekNBigInt(1, e)
	if index.Sign() < 0 {
		log.Error("[validatorSetItem] index < 0")
		return ErrBadValue
	}
	arrItem := PeekN(2, e)
	if arrItem == nil {
		log.Error("[validatorSetItem] arrItem = nil")
		return ErrBadValue
	}
	item := arrItem.GetStackItem()
	if _, ok := item.(*types.Array); !ok {
		if _, ok := item.(*types.ByteArray); ok {
			l := len(item.GetByteArray())
			if index.Cmp(big.NewInt(int64(l))) >= 0 {
				log.Error("[validatorSetItem] index >= l")
				return ErrOverMaxArraySize
			}
			if len(newItem.GetStackItem().GetByteArray()) == 0 {
				log.Error("[validatorSetItem] len(newItem.GetStackItem().GetByteArray()) = 0 ")
				return ErrBadValue
			}
		} else {
			log.Error("[validatorSetItem] ErrBadValue")
			return ErrBadValue
		}
	} else {
		if index.Cmp(big.NewInt(int64(len(item.GetArray())))) >= 0 {
			log.Error("[validatorSetItem] index >= len(item.GetArray())")
			return ErrOverMaxArraySize
		}
	}
	return nil
}

func validateNewArray(e *ExecutionEngine) error {
	if err := LogStackTrace(e, 1, "[validateNewArray]"); err != nil {
		return err
	}

	count := PeekBigInteger(e)
	if count.Sign() < 0 {
		return ErrBadValue
	}
	if count.Cmp(big.NewInt(int64(MaxArraySize))) > 0 {
		log.Error("[validateNewArray] uint32(count) > MaxArraySize ")
		return ErrOverMaxArraySize
	}
	return nil
}

func CheckBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(types.ConvertBigIntegerToBytes(value)) > MaxSizeForBigInteger {
		return false
	}
	return true
}

func LogStackTrace(e *ExecutionEngine, needStackCount int, desc string) error {
	stackCount := EvaluationStackCount(e)
	if stackCount < needStackCount {
		log.Error(fmt.Sprintf("%s lack of parametes, actual: %v need: %x", desc, stackCount, needStackCount))
		return ErrUnderStackLen
	}
	return nil
}
