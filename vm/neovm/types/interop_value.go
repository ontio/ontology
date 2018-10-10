package types

import (
	"bytes"
	"github.com/ontio/ontology/vm/neovm/interfaces"
)

type InteropValue struct {
	data interfaces.Interop
}

func NewInteropValue(value interfaces.Interop) InteropValue {
	return InteropValue{data: value}
}

func (this *InteropValue) Equals(other InteropValue) bool {
	// todo: both nil?
	if this.data == nil || other.data == nil {
		return false
	}
	return bytes.Equal(this.data.ToArray(), other.data.ToArray())
}

