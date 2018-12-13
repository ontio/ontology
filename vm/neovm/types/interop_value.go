package types

import (
	"bytes"
	"github.com/ontio/ontology/vm/neovm/interfaces"
)

type InteropValue struct {
	Data interfaces.Interop
}

func NewInteropValue(value interfaces.Interop) InteropValue {
	return InteropValue{Data: value}
}

func (this *InteropValue) Equals(other InteropValue) bool {
	// todo: both nil?
	if this.Data == nil || other.Data == nil {
		return false
	}
	return bytes.Equal(this.Data.ToArray(), other.Data.ToArray())
}
