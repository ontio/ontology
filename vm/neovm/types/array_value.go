package types

import (
	"fmt"

	"github.com/ontio/ontology/vm/neovm/constants"
	"github.com/ontio/ontology/vm/neovm/errors"
)

type ArrayValue struct {
	Data []VmValue
}

const initArraySize = 16

func NewArrayValue() *ArrayValue {
	return &ArrayValue{Data: make([]VmValue, 0, initArraySize)}
}

func (self *ArrayValue) Append(item VmValue) error {
	if len(self.Data) >= constants.MAX_ARRAY_SIZE {
		return errors.ERR_OVER_MAX_ARRAY_SIZE
	}
	self.Data = append(self.Data, item)
	return nil
}

func (self *ArrayValue) Len() int64 {
	return int64(len(self.Data))
}

func (self *ArrayValue) RemoveAt(index int64) error {
	if index < 0 || index >= self.Len() {
		return fmt.Errorf("[RemoveAt] index out of bound!")
	}
	self.Data = append(self.Data[:index], self.Data[index+1:]...)
	return nil
}
