package neovm

import (
	"bytes"
	"encoding/binary"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/vm/neovm/types"
)

func validateCount1(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 1 {
		log.Error("[validateCount1]", EvaluationStackCount(e) < 1)
		return ErrUnderStackLen
	}
	return nil
}

func validateCount2(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validateCount2]", EvaluationStackCount(e) < 2)
		return ErrUnderStackLen
	}
	return nil
}

func validateCount3(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 3 {
		log.Error("[validateCount3]", EvaluationStackCount(e) < 3)
		return ErrUnderStackLen
	}
	return nil
}

func validateDivMod(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validateDivMod]", EvaluationStackCount(e) < 2)
		return ErrUnderStackLen
	}
	if PeekInt(e) == 0 {
		return ErrDivModByZero
	}
	return nil
}

func validatorPushData4(e *ExecutionEngine) error {
	index := e.context.GetInstructionPointer()
	if index + 4 >= len(e.context.Code) {
		return ErrOverCodeLen
	}
	bytesBuffer := bytes.NewBuffer(e.context.Code[index : index + 4])
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
	if uint32(e.invocationStack.Count()) > MaxStackSize {
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

func validateOpStack(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 1 {
		log.Error("[validateOpStack]", EvaluationStackCount(e) < 1)
		return ErrUnderStackLen
	}
	index := PeekNInt(0, e)
	if index < 0 {
		log.Error("[validateOpStack] index < 0")
		return ErrBadValue
	}

	return nil
}

func validateXDrop(e *ExecutionEngine) error {
	if err := validateOpStack(e); err != nil {
		return err
	}
	return nil
}

func validateXSwap(e *ExecutionEngine) error {
	if err := validateOpStack(e); err != nil {
		return err
	}
	return nil
}

func validateXTuck(e *ExecutionEngine) error {
	if err := validateOpStack(e); err != nil {
		return err
	}
	return nil
}

func validatePick(e *ExecutionEngine) error {
	if err := validateOpStack(e); err != nil {
		return err
	}
	return nil
}

func validateRoll(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 1 {
		log.Error("[validateRoll]", EvaluationStackCount(e) < 1)
		return ErrUnderStackLen
	}
	index := PeekNInt(0, e)
	if index < 0 {
		log.Error("[validateRoll] index < 0")
		return ErrBadValue
	}
	return nil
}

func validateCat(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validateCat] EvaluationStackCount(e) < 2")
		return ErrUnderStackLen
	}
	l := len(PeekNByteArray(0, e)) + len(PeekNByteArray(1, e))
	if uint32(l) > MaxItemSize {
		log.Error("[validateCat] uint32(l) > MaxItemSize")
		return ErrOverMaxItemSize
	}
	return nil
}

func validateSubStr(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 3 {
		log.Error("[validateSubStr]", EvaluationStackCount(e) < 3)
		return ErrUnderStackLen
	}
	count := PeekNInt(0, e)
	if count < 0 {
		log.Error("[validateSubStr] count < 0")
		return ErrBadValue
	}
	index := PeekNInt(1, e)
	if index < 0 {
		log.Error("[validateSubStr] index < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(2, e)
	if len(arr) < index + count {
		log.Error("[validateSubStr] len(arr) < index + count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validateLeft(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validateLeft]", EvaluationStackCount(e) < 2)
		return ErrUnderStackLen
	}
	count := PeekNInt(0, e)
	if count < 0 {
		log.Error("[validateLeft] count < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(1, e)
	if len(arr) < count {
		log.Error("[validateLeft] len(arr) < count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validateRight(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validateRight]", EvaluationStackCount(e) < 2)
		return ErrUnderStackLen
	}
	count := PeekNInt(0, e)
	if count < 0 {
		log.Error("[validateRight] count < 0")
		return ErrBadValue
	}
	arr := PeekNByteArray(1, e)
	if len(arr) < count {
		log.Error("[validateRight] len(arr) < count")
		return ErrOverMaxArraySize
	}
	return nil
}

func validatePack(e *ExecutionEngine) error {
	count := PeekInt(e)
	if uint32(count) > MaxArraySize {
		log.Error("[validateRight] uint32(count) > MaxArraySize")
		return ErrOverMaxArraySize
	}
	if count > EvaluationStackCount(e) {
		log.Error("[validateRight] count > EvaluationStackCount(e)")
		return ErrOverStackLen
	}
	return nil
}

func validatePickItem(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 2 {
		log.Error("[validatePickItem]", EvaluationStackCount(e) < 2)
		return ErrUnderStackLen
	}
	index := PeekNInt(0, e)
	if index < 0 {
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
	if index >= len(stackItem.GetArray()) {
		log.Error("[validatePickItem] index >= len(stackItem.GetArray())")
		return ErrOverMaxArraySize
	}
	return nil
}

func validatorSetItem(e *ExecutionEngine) error {
	if EvaluationStackCount(e) < 3 {
		log.Error("[validatorSetItem]", EvaluationStackCount(e) < 3)
		return ErrUnderStackLen
	}
	newItem := PeekN(0, e)
	if newItem == nil {
		log.Error("[validatorSetItem] newItem = nil")
		return ErrBadValue
	}
	index := PeekNInt(1, e)
	if index < 0 {
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
			if index >= l {
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
		if index >= len(item.GetArray()) {
			log.Error("[validatorSetItem] index >= len(item.GetArray())")
			return ErrOverMaxArraySize
		}
	}
	return nil
}

func validateNewArray(e *ExecutionEngine) error {
	count := PeekInt(e)
	if uint32(count) > MaxArraySize {
		log.Error("[validateNewArray] uint32(count) > MaxArraySize ")
		return ErrOverMaxArraySize
	}
	return nil
}
