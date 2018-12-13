package types

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSerialize(t *testing.T) {

	bsValue, err := VmValueFromBytes([]byte("test"))
	assert.Equal(t, err, nil)
	boolValue := VmValueFromBool(true)
	bigin := new(big.Int)
	bigin.SetInt64(int64(1000))
	biginValue, err := VmValueFromBigInt(bigin)
	assert.Equal(t, err, nil)
	uint64Value := VmValueFromUint64(uint64(100))
	s := NewStructValue()
	s = s.Append(bsValue)
	s = s.Append(boolValue)
	s = s.Append(biginValue)
	s = s.Append(uint64Value)
	structValue := VmValueFromStructVal(s)
	sink := new(common.ZeroCopySink)
	structValue.Serialize(sink)
	fmt.Println(common.ToHexString(sink.Bytes()))
	assert.Equal(t, common.ToHexString(sink.Bytes()), "810400047465737401010202e803020164")

	structValueStr, err := structValue.ConvertNeoVmValueHexString()
	fmt.Println("structValueStr:", structValueStr)

	source := common.NewZeroCopySource(sink.Bytes())
	vs := VmValue{}
	vs.Deserialize(source)
	assert.Equal(t, structValue, vs)

	arr := NewArrayValue()
	arr.Append(bsValue)
	arr.Append(boolValue)
	arr.Append(biginValue)
	arr.Append(uint64Value)
	sinkArr := new(common.ZeroCopySink)
	arrValue := VmValueFromArrayVal(arr)
	arrValue.Serialize(sinkArr)
	assert.Equal(t, common.ToHexString(sinkArr.Bytes()), "800400047465737401010202e803020164")

	arrValue2 := VmValue{}
	source = common.NewZeroCopySource(sinkArr.Bytes())
	arrValue2.Deserialize(source)
	assert.Equal(t, arrValue2, arrValue)

	m := NewMapValue()
	m.Set(bsValue, arrValue)
	m.Set(biginValue, structValue)
	m.Set(uint64Value, boolValue)

	//map test have problem
	//TODO
	mValue := VmValueFromMapValue(m)
	sinkMap := new(common.ZeroCopySink)
	mValue.Serialize(sinkMap)
	//assert.Equal(t, common.ToHexString(sinkMap.Bytes()), "82030201640101000474657374800400047465737401010202e8030201640202e803810400047465737401010202e803020164")
}
