package neovm

import "github.com/ontio/ontology/vm/neovm/types"

func (self *ValueStack) PushBool(val bool) error {
	if val {
		return self.Push(types.VmValueFromInt64(1))
	} else {
		return self.Push(types.VmValueFromInt64(0))
	}
}

func (self *ValueStack) PopAsBool() (bool, error) {
	val, err := self.Pop()
	if err != nil {
		return false, err
	}

	return val.AsBool()
}

func (self *ValueStack) PushInt64(val int64) error {
	return self.Push(types.VmValueFromInt64(val))
}

func (self *ValueStack) PopAsInt64() (int64, error) {
	val, err := self.Pop()
	if err != nil {
		return 0, err
	}
	return val.AsInt64()
}

func (self *ValueStack) PopAsIntValue() (types.IntValue, error) {
	val, err := self.Pop()
	if err != nil {
		return types.IntValue{}, err
	}
	return val.AsIntValue()
}

func (self *ValueStack) PushBytes(val []byte) error {
	v, err := types.VmValueFromBytes(val)
	if err != nil {
		return err
	}
	return self.Push(v)
}

func (self *ValueStack) PopAsBytes() ([]byte, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsBytes()
}

func (self *ValueStack) PopAsArray() (*types.ArrayValue, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsArrayValue()
}

func (self *ValueStack) PopAsMap() (*types.MapValue, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsMapValue()
}

func (self *ValueStack) PopAsStruct() (types.StructValue, error) {
	val, err := self.Pop()
	if err != nil {
		return types.StructValue{}, err
	}
	return val.AsStructValue()
}

func (self *ValueStack) PopPairAsBytes() (left, right []byte, err error) {
	right, err = self.PopAsBytes()
	if err != nil {
		return
	}
	left, err = self.PopAsBytes()
	return
}

func (self *ValueStack) PopPairAsBool() (left, right bool, err error) {
	right, err = self.PopAsBool()
	if err != nil {
		return
	}
	left, err = self.PopAsBool()
	return
}

func (self *ValueStack) PopPairAsInt64() (left, right int64, err error) {
	right, err = self.PopAsInt64()
	if err != nil {
		return
	}
	left, err = self.PopAsInt64()
	return
}

func (self *ValueStack) PopPairAsIntVal() (left, right types.IntValue, err error) {
	right, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	left, err = self.PopAsIntValue()
	return
}

func (self *ValueStack) PopTripleAsBytes() (left, middle, right []byte, err error) {
	right, err = self.PopAsBytes()
	if err != nil {
		return
	}
	middle, err = self.PopAsBytes()
	if err != nil {
		return
	}
	left, err = self.PopAsBytes()
	return
}

func (self *ValueStack) PopTripleAsBool() (left, middle, right bool, err error) {
	right, err = self.PopAsBool()
	if err != nil {
		return
	}
	middle, err = self.PopAsBool()
	if err != nil {
		return
	}
	left, err = self.PopAsBool()
	return
}

func (self *ValueStack) PopTripleAsIntVal() (left, middle, right types.IntValue, err error) {
	right, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	middle, err = self.PopAsIntValue()
	if err != nil {
		return
	}
	left, err = self.PopAsIntValue()
	return
}
