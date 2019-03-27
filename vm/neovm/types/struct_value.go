package types

import "fmt"

// struct value is value type
type StructValue struct {
	Data []VmValue
}

func NewStructValue() *StructValue {
	return &StructValue{Data: make([]VmValue, 0, initArraySize)}
}

func (self *StructValue) Append(item VmValue) {
	self.Data = append(self.Data, item)
}

func (self *StructValue) Len() int64 {
	return int64(len(self.Data))
}

func (self *StructValue) Clone() (*StructValue, error) {
	var i int
	return cloneStruct(self, &i)
}

func cloneStruct(s *StructValue, length *int) (*StructValue, error) {
	if *length > MAX_CLONE_LENGTH {
		return nil, fmt.Errorf("%s", "over max struct clone length")
	}
	var arr []VmValue
	for _, v := range s.Data {
		*length++
		if temp, err := v.AsStructValue(); err == nil {
			vc, err := cloneStruct(temp, length)
			if err != nil {
				return nil, err
			}
			arr = append(arr, VmValueFromStructVal(vc))
		} else {
			arr = append(arr, v)
		}
	}
	return &StructValue{Data: arr}, nil
}
