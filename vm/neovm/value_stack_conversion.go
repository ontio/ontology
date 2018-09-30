package neovm

import "github.com/ontio/ontology/vm/neovm/types"

func (self *ValueStack) PopAsBool() (bool, error) {
	val, err := self.Pop()
	if err != nil {
		return false, err
	}

	return val.AsBool()
}

func (self *ValueStack) PopAsIntValue() (types.IntValue, error) {
	val, err := self.Pop()
	if err != nil {
		return types.IntValue{}, err
	}
	return val.AsIntValue()
}

func (self *ValueStack) PopAsBytes() ([]byte, error) {
	val, err := self.Pop()
	if err != nil {
		return nil, err
	}
	return val.AsBytes()
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
