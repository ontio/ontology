package utils

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestUnexportFields(t *testing.T) {
	unexport := struct {
		Name string
		num  uint
		Age  int
	}{
		Name: "aaa",
		num:  100,
		Age:  123,
	}
	export := struct {
		Name string
		Age  int
	}{
		Name: "aaa",
		Age:  123,
	}

	unexportCode, err := BuildNeoVMInvokeCode(common.Address{}, []interface{}{unexport})
	assert.Nil(t, err)
	exportCode, err := BuildNeoVMInvokeCode(common.Address{}, []interface{}{export})
	assert.Nil(t, err)
	assert.Equal(t, unexportCode, exportCode)
}
