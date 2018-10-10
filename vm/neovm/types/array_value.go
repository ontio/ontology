package types

type ArrayValue struct {
	data []VmValue
}

const initArraySize = 16

func NewArrayValue() *ArrayValue {
	return &ArrayValue{data: make([]VmValue, 0, initArraySize)}
}

func (self *ArrayValue) Append(item VmValue) {
	self.data = append(self.data, item)
}

