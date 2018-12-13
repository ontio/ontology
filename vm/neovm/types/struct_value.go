package types

// struct value is value type
type StructValue struct {
	Data []VmValue
}

func NewStructValue() StructValue {
	return StructValue{Data: make([]VmValue, 0, initArraySize)}
}

func (self *StructValue) Append(item VmValue) StructValue {
	return StructValue{Data: append(self.Data, item)}
}

func (self *StructValue) Len() int64 {
	return int64(len(self.Data))
}
