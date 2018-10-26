package types

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

func (self *ArrayValue)Len() int64 {
	return int64(len(self.Data))
}

