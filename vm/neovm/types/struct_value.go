package types

// struct value is value type
type StructValue struct {
	data []VmValue
}

func (self *StructValue) Append(item VmValue) StructValue {
	return StructValue{data: append(self.data, item)}
}
