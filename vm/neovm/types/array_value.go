package types

import "fmt"

type ArrayValue struct {
	Data []VmValue
}

const initArraySize = 16

func NewArrayValue() *ArrayValue {
	return &ArrayValue{Data: make([]VmValue, 0, initArraySize)}
}

func (self *ArrayValue) Append(item VmValue) {
	//todo: check limit
	self.Data = append(self.Data, item)
}

func (self *ArrayValue) Len() int64 {
	return int64(len(self.Data))
}

func (self *ArrayValue)RemoveAt(index int64) error {
	if index >= self.Len() {
		return fmt.Errorf("[RemoveAt] index out of bound!")
	}
	self.Data = append(self.Data[:index-1], self.Data[index:]...)
	return nil
}
